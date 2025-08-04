package command

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/qoinlyid/qore"
	"golang.org/x/term"
)

// Prompt displays prompt command and return `string` otherwise `error`
func (c *command) Prompt(question string, required ...bool) (*Output[string], error) {
	var reqd bool
	if len(required) > 0 {
		reqd = required[0]
	}
	if !reqd {
		question = c.questioner(question, "Optional")
	} else {
		question = c.questioner(question)
	}

	// Use raw terminal mode untuk konsistensi dengan choice/multichoice
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, errors.Join(ErrMakeTermRaw, err)
	}
	defer term.Restore(fd, oldState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		term.Restore(fd, oldState)
		os.Exit(1)
	}()

	clearScreen()
	fmt.Fprint(c.writer, question)

	// Read input character by character
	var input strings.Builder
	b := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			continue
		}

		key := b[0]

		switch key {
		case 3: // Ctrl+C
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrInterrupted
		case 27: // ESC key
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrCancelled
		case 13, 10: // Enter key (CR or LF)
			text := input.String()
			fmt.Fprint(c.writer, "\r\n")

			if reqd && qore.ValidationIsEmpty(text) {
				return nil, ErrResponseRequired
			}

			return newOutput(question, text), nil
		case 127, 8: // Backspace key
			if input.Len() > 0 {
				// Remove last character from input
				currentInput := input.String()
				input.Reset()
				input.WriteString(currentInput[:len(currentInput)-1])

				// Move cursor back, write space, move back again
				fmt.Fprint(c.writer, "\b \b")
			}
		default:
			// Only accept printable characters
			if key >= 32 && key <= 126 {
				input.WriteByte(key)
				fmt.Fprintf(c.writer, "%c", key)
			}
		}
	}
}

// PromptWithDefault displays prompt command with default value and return `string`.
func (c *command) PromptWithDefault(question string, defaultValue string) (*Output[string], error) {
	question = c.questioner(question, fmt.Sprintf("Default: %s", defaultValue))

	// Use raw terminal mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, errors.Join(ErrMakeTermRaw, err)
	}
	defer term.Restore(fd, oldState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		term.Restore(fd, oldState)
		os.Exit(1)
	}()

	clearScreen()
	fmt.Fprint(c.writer, question)

	// Read input character by character
	var input strings.Builder
	b := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			continue
		}

		key := b[0]

		switch key {
		case 3: // Ctrl+C
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrInterrupted
		case 27: // ESC key
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrCancelled
		case 13, 10: // Enter key (CR or LF)
			text := input.String()
			fmt.Fprint(c.writer, "\r\n")

			if qore.ValidationIsEmpty(text) {
				return newOutput(question, defaultValue), nil
			}

			return newOutput(question, text), nil
		case 127, 8: // Backspace key
			if input.Len() > 0 {
				// Remove last character from input
				currentInput := input.String()
				input.Reset()
				input.WriteString(currentInput[:len(currentInput)-1])

				// Move cursor back, write space, move back again
				fmt.Fprint(c.writer, "\b \b")
			}
		default:
			// Only accept printable characters
			if key >= 32 && key <= 126 {
				input.WriteByte(key)
				fmt.Fprintf(c.writer, "%c", key)
			}
		}
	}
}

package command

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/term"
)

func (c *command) ConfirmWithDefault(question string, defaultValue bool) (*Output[bool], error) {
	defaultText := "Y/n"
	if !defaultValue {
		defaultText = "y/N"
	}
	question = c.questioner(question, defaultText)
	fmt.Fprint(c.writer, question)

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

	// Read input.
	b := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			continue
		}

		key := b[0]
		var result bool
		var resultText string

		switch key {
		case 'y', 'Y':
			result = true
			resultText = "yes"
		case 'n', 'N':
			result = false
			resultText = "no"
		case 13, 10: // Enter key - use default
			result = defaultValue
			if defaultValue {
				resultText = "yes"
			} else {
				resultText = "no"
			}
		case 3: // Ctrl+C
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrInterrupted
		case 27: // ESC key
			fmt.Fprint(c.writer, "\r\n")
			return nil, ErrCancelled
		default:
			continue // Invalid input, continue reading
		}

		fmt.Fprintf(c.writer, "%s\r\n", resultText)
		return newOutput(question, result), nil
	}
}

func (c *command) Confirm(question string) (*Output[bool], error) {
	return c.ConfirmWithDefault(question, true)
}

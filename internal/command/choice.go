package command

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/term"
)

// Option defines option argument for command choice.
type Option struct {
	Key   any
	Value string
}

// Choice displays the single choice prompt and returns the `*Output` otherwise `error`.
func (c *command) Choice(question string, options []Option, required ...bool) (*Output[Option], error) {
	var reqd bool
	if len(required) > 0 {
		reqd = required[0]
	}
	if !reqd {
		question = c.questioner(question, "Optional")
	} else {
		question = c.questioner(question)
	}
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

	index := 0
	selected := -1

	// Anonymous render function for single choice
	render := func() {
		clearScreen()
		fmt.Fprint(c.writer, question)
		fmt.Fprint(c.writer, "\r\n")
		for i, opt := range options {
			prefix := "( )"
			if selected == i {
				prefix = "(o)"
			}
			if i == index {
				fmt.Fprint(c.writer, " > ")
			} else {
				fmt.Fprint(c.writer, "   ")
			}
			fmt.Fprint(c.writer, prefix)
			fmt.Fprint(c.writer, " ")
			fmt.Fprint(c.writer, opt.Value)
			fmt.Fprint(c.writer, "\r\n")
		}
		fmt.Fprint(c.writer, "\r\n")
		fmt.Fprint(c.writer, "Use arrow keys to navigate, space to select/deselect, enter to confirm")
		fmt.Fprint(c.writer, "\r\n")
	}

	// Anonymous reader function
	read := func() byte {
		b := make([]byte, 3)
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			return 0
		}

		// Handle escape sequences (arrow keys)
		if n >= 3 && b[0] == 27 && b[1] == 91 {
			return b[2] // arrow keys
		}
		return b[0]
	}

	render()
	for {
		key := read()
		switch key {
		case 65: // Up arrow key
			if index > 0 {
				index--
			}
		case 66: // Down arrow key
			if index < len(options)-1 {
				index++
			}
		case 32: // Spacebar key (toggle)
			if selected == index {
				selected = -1 // Deselect if already selected
			} else {
				selected = index // Select current option
			}
		case 13, 10: // Enter key (CR or LF)
			clearScreen()
			if reqd && selected == -1 {
				return nil, ErrResponseRequired
			}
			return newOutput(question, options[index]), nil
		case 3: // Ctrl+C
			clearScreen()
			return nil, ErrInterrupted
		case 27: // ESC key
			clearScreen()
			return nil, ErrCancelled
		}
		render()
	}
}

// MultiChoice displays the multiple choice prompt and returns the `*Output` otherwise `error`.
func (c *command) MultiChoice(question string, options []Option, required ...bool) (*Output[[]Option], error) {
	var reqd bool
	if len(required) > 0 {
		reqd = required[0]
	}
	if !reqd {
		question = c.questioner(question, "Optional")
	} else {
		question = c.questioner(question)
	}
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

	selecteds := make([]bool, len(options))
	index := 0

	// Anonymous render function.
	render := func() {
		clearScreen()
		fmt.Fprint(c.writer, question)
		fmt.Fprint(c.writer, "\r\n")
		for i, opt := range options {
			prefix := "[ ]"
			if selecteds[i] {
				prefix = "[✓]"
			}
			if i == index {
				fmt.Fprint(c.writer, " > ")
				fmt.Fprint(c.writer, prefix)
				fmt.Fprint(c.writer, " ")
				fmt.Fprint(c.writer, opt.Value)
			} else {
				fmt.Fprint(c.writer, "   ")
				fmt.Fprint(c.writer, prefix)
				fmt.Fprint(c.writer, " ")
				fmt.Fprint(c.writer, opt.Value)
			}
			fmt.Fprint(c.writer, "\r\n")
		}
		fmt.Fprint(c.writer, "\r\n")
		fmt.Fprint(c.writer, "Use arrow keys to navigate, space to select/deselect, enter to confirm")
		fmt.Fprint(c.writer, "\r\n")
	}

	// Anonymous reader function with better key detection.
	read := func() byte {
		b := make([]byte, 3)
		n, err := os.Stdin.Read(b)
		if err != nil || n == 0 {
			return 0
		}

		// Handle escape sequences (arrow keys)
		if n >= 3 && b[0] == 27 && b[1] == 91 {
			return b[2] // arrow keys
		}
		return b[0]
	}

	render()
	for {
		key := read()
		switch key {
		case 65: // Up arrow key (or 'A')
			if index > 0 {
				index--
			}
		case 66: // Down arrow key (or 'B')
			if index < len(options)-1 {
				index++
			}
		case 32: // Spacebar key (toggle)
			selecteds[index] = !selecteds[index]
		case 13, 10: // Enter key (CR or LF)
			var result []Option
			for i, selected := range selecteds {
				if selected {
					result = append(result, options[i])
				}
			}
			clearScreen()
			if reqd && len(result) == 0 {
				return nil, ErrResponseRequired
			}
			return newOutput(question, result), nil
		case 3: // Ctrl+C
			clearScreen()
			return nil, ErrInterrupted
		case 27: // ESC key
			clearScreen()
			return nil, ErrCancelled
		}
		render()
	}
}

package command

import (
	"fmt"
	"os"
)

func clearScreen() {
	fmt.Print("\033[2J")   // Clear entire screen.
	fmt.Print("\033[H")    // Move cursor to top-left.
	fmt.Print("\033[0;0H") // Alternative cursor positioning.

	// Force flush the output.
	os.Stdout.Sync()
}

func clearTerminalHistory() {
	// Clear scrollback buffer (works on most modern terminals).
	fmt.Print("\033[3J") // Clear scrollback buffer.
	fmt.Print("\033[2J") // Clear entire screen.
	fmt.Print("\033[H")  // Move cursor to home position.

	// Alternative method for some terminals.
	fmt.Print("\033c") // Reset terminal (ESC c).

	// Force flush.
	os.Stdout.Sync()
}

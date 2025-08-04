package command

// ResultColor defines available colors for result text.
type ResultColor uint8

const (
	ColorDefault   ResultColor = iota // Auto-detect or default.
	ColorLightGray                    // Light gray (\033[37m).
	ColorWhite                        // White (\033[97m).
	ColorCyan                         // Cyan (\033[96m).
	ColorLightBlue                    // Light blue (\033[94m).
	ColorYellow                       // Yellow (\033[93m).
	ColorGreen                        // Green (\033[92m).
)

// getColorCode returns ANSI color code for ResultColor.
func (c ResultColor) getColorCode() string {
	switch c {
	case ColorLightGray:
		return "\033[37m"
	case ColorWhite:
		return "\033[97m"
	case ColorCyan:
		return "\033[96m"
	case ColorLightBlue:
		return "\033[94m"
	case ColorYellow:
		return "\033[93m"
	case ColorGreen:
		return "\033[92m"
	default: // ColorDefault
		return "\033[96m" // Default to light cyan.
	}
}

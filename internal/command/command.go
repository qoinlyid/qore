package command

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// Config defines command configuration.
type Config struct {
	QuestionMark  string
	PrintResponse bool
}

// command defines command instance.
type command struct {
	config *Config
	reader *bufio.Reader
	writer io.Writer
}

// defaultConfig for command.
var defaultConfig = &Config{
	QuestionMark:  "?",
	PrintResponse: true,
}

// NewCommand returns command instance.
func NewCommand(config ...Config) *command {
	// Config.
	cfg := defaultConfig
	if len(config) > 0 {
		if config[0] != (Config{}) {
			cfg = &config[0]
		}
	}
	if cfg.QuestionMark == "" {
		cfg.QuestionMark = defaultConfig.QuestionMark
	}

	// Command.
	enableVirtualTerminalProcessing()
	return &command{
		config: cfg,
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// questioner sets the standarize command question.
func (c *command) questioner(question string, suffix ...string) string {
	if !strings.HasSuffix(question, c.config.QuestionMark) {
		question = question + c.config.QuestionMark
	}
	if len(suffix) > 0 {
		if suffix[0] != "" {
			question = question + " (" + suffix[0] + ")"
		}
	}
	return question + " "
}

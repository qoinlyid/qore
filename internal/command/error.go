package command

import "errors"

var (
	// System.
	ErrReaderFailedRead = errors.New("command reader failed to read")
	ErrMakeTermRaw      = errors.New("failed to create raw terminal")

	// Prompt.
	ErrCancelled        = errors.New("cancelled by user")
	ErrInterrupted      = errors.New("interrupted by user")
	ErrResponseRequired = errors.New("response is required")
)

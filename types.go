package qore

import (
	"os"
)

// Supervisor is process manager that manage application process.
type Supervisor interface {
	// ListenSignals returns signal(s) that must be listened to do gracefully stop.
	ListenSignals() []os.Signal
	// Run is the main method responsible for running the application.
	Run(app *App)
}

// Numeric custom type.
type Numeric interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

// Validatable custom type.
type Validatable interface {
	Numeric | []any | map[string]string | map[string]any |
		map[any]any | string | any |
		[]string
}

// LogLevel custom type for logging level.
type LogLevel string

// Dependency is package that will be depent on module(s).
type Dependency interface {
	Name() string
	Priority() int
	IsReady() bool
	Open() error
	Close() error
}

// Module is qore module interface.
type Module interface {
	HttpRoutes(app *App)
}

// ModuleLoader is module loader interface.
type ModuleLoader interface {
	Load() []Module
}

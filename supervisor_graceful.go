package qore

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"
)

// SupervisorGraceful implement Supervisor and manage the application gracefully.
type SupervisorGraceful struct {
	// Signals that must be listened to do gracefully stop.
	Signals []os.Signal
	// HookStart will be executed before the application start.
	HookStart func()
	// HookStop will be executed before the application stop.
	HookStop func()
	// BinaryFilePath location of the binary file.
	BinaryFilePath string
	// Graceful shutdown or restart signal.
	GracefulRestartSignal os.Signal
	// Graceful shutdown or restart interval seconds.
	GracefulInterval int

	// Private field.
	app *App
}

// Compile time check SupervisorGraceful implements Supervisor.
var _ Supervisor = (*SupervisorGraceful)(nil)

func (s *SupervisorGraceful) ListenSignals() []os.Signal {
	return s.Signals
}
func (s *SupervisorGraceful) Run(app *App) {
	// Default modifier.
	s.app = app
	if s.GracefulRestartSignal == nil {
		s.GracefulRestartSignal = syscall.SIGUSR2
	}
	if s.GracefulInterval <= 0 || s.GracefulInterval > 12 {
		s.GracefulInterval = 12
	}
	if len(s.Signals) == 0 {
		s.Signals = []os.Signal{
			syscall.SIGUSR2,
			syscall.SIGHUP,
			syscall.SIGTSTP,
			syscall.SIGINT,
			syscall.SIGKILL,
			os.Interrupt,
		}
	}

	// Run under overseer.
	overseer.Run(overseer.Config{
		Program:       s.program,
		Addresses:     app.addresses,
		Debug:         !app.Config.AppProduction,
		RestartSignal: s.GracefulRestartSignal,
		Fetcher: &fetcher.File{
			Path:     s.BinaryFilePath,
			Interval: time.Second * time.Duration(s.GracefulInterval),
		},
	})
}

func (s *SupervisorGraceful) program(state overseer.State) {
	// Execute hook function before start if any.
	if s.HookStart != nil {
		s.HookStart()
	}

	// Start running the application server.
	s.app.startServer(state.Listeners...)

	// Signal notify.
	ctx, stop := signal.NotifyContext(context.Background(), s.Signals...)
	defer stop()

	// Terminate.
	<-ctx.Done()

	// Execute hook function before stop if any.
	if s.HookStop != nil {
		s.HookStop()
	}

	// Stop the running application server.
	// Question: is needed to stop tcp listener while using overseer?
}

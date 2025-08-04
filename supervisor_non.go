package qore

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SupervisorNon implement Supervisor but does not manage the application restart gracefully.
type SupervisorNon struct {
	// Signals that must be listened to do gracefully stop.
	Signals []os.Signal
	// HookStart will be executed before the application start.
	HookStart func()
	// HookStop will be executed before the application stop.
	HookStop func()
}

// Compile time check SupervisorNon implements Supervisor.
var _ Supervisor = (*SupervisorNon)(nil)

func (s *SupervisorNon) ListenSignals() []os.Signal {
	return s.Signals
}
func (s *SupervisorNon) Run(app *App) {
	// Default modifier.
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

	// Execute hook function before start if any.
	if s.HookStart != nil {
		s.HookStart()
	}

	// Start running the application server.
	app.startServer()

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
	app.stopServer()
}

package qore

import (
	"fmt"
	"net"
)

func (app *App) startServer(listeners ...net.Listener) {
	logger := app.Logger().Group("lifecycle.start")

	// Given slice listener, the net listener should be handled by supervisor.
	if len(listeners) > 0 {
		for i, listener := range listeners {
			if listener == nil {
				continue
			}

			// Index 0 must be HTTP server listener.
			if i == 0 && app.httpServer != nil {
				app.httpServer.start(func() (listener net.Listener, err error) {
					return listener, nil
				}, logger)
			}
		}
		return
	}

	// Empty slice of listener from supervisor, so should be handled manually.
	// HTTP server.
	if app.httpServer != nil {
		app.httpServer.start(func() (listener net.Listener, err error) {
			// Resolve TCP.
			address := fmt.Sprintf(":%d", app.Config.HTTPPort)
			addr, err := net.ResolveTCPAddr("tcp", address)
			if err != nil {
				return nil, fmt.Errorf("failed resolve listen %s: %w", address, err)
			}

			// Listen TCP.
			listener, err = net.ListenTCP("tcp", addr)
			if err != nil {
				return nil, fmt.Errorf("failed to listen TCP %s: %w", address, err)
			}
			return
		}, logger)
	}
}

func (app *App) stopServer() {
	logger := app.Logger().Group("lifecycle.stop")

	// Stoping HTTP server.
	if app.httpServer != nil {
		app.httpServer.stop(logger)
	}
}

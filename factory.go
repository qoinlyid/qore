package qore

import (
	"fmt"
	"time"
)

// New creates an application services.
//
//	app := qore.New()
func New() *App {
	app := new(App)

	// Config load.
	app.Config = loadConfig()
	if app.Config.HTTPPort > 0 {
		app.addresses = append(app.addresses, fmt.Sprintf(":%d", app.Config.HTTPPort))
	}

	// Utility
	app.logger = setupLogger(app.Config)

	// Application server.
	if app.Config.HTTPPort > 0 {
		// HTTP server.
		app.httpServer = newHttpServer()
		app.httpServer.autoTLS = app.Config.HTTPAutoTLS
		app.httpServer.certPath = app.Config.HTTPCertPath
		app.httpServer.keyPath = app.Config.HTTPKeyPath
		app.httpServer.shutdownTimeout = time.Duration(app.Config.ShutdownTimeout) * time.Second
		app.httpServer.core.Debug = !app.Config.AppProduction
	}

	return app
}

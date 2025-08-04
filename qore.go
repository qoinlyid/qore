// Qore Go toolkit.
package qore

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"syscall"
)

// App is the Qore main application.
type App struct {
	// Config value that parsed from env or file, can be used by application.
	Config *Config

	// Unexported utility.
	logger *logger

	// Unexported application server.
	httpServer *httpServer

	// Unexported application addresses that can used by supervisor.
	addresses []string

	// Unexported dependency registry
	dependencyRegistry []Dependency
}

// Logger instance that associated with the app.
func (a *App) Logger() *logger {
	return a.logger
}

// SetHttpMiddleware will set given middleware(s) to the global HTTP(S) middleware.
func (app *App) SetHttpMiddleware(middlewares ...HttpMiddleware) {
	if app.httpServer == nil {
		return
	}
	app.httpServer.core.Use(httpMiddlewareWrappers(app.httpServer, app.logger, middlewares...)...)
}

// SetHttpRoutes creates HTTP routers object that can be used for registering HTTP route.
func (app *App) SetHttpRoutes(fn func(router *HttpRouter)) {
	if app.httpServer == nil {
		return
	}
	router := &HttpRouter{server: app.httpServer, logger: app.logger}
	fn(router)
}

// SetHttpValidator will set custom request validator to the HTTP(S).
func (app *App) SetHttpValidator(validator HttpValidator) {
	if app.httpServer == nil {
		return
	} else if validator == nil {
		return
	}
	app.httpServer.validator = validator
}

// SetApiResponseInterface will set custom HTTP(s) API response wrapper.
func (app *App) SetApiResponseInterface(iApiResponse ApiResponseInterface) {
	if app.httpServer == nil {
		return
	} else if iApiResponse == nil {
		return
	}
	app.httpServer.iApiResponse = iApiResponse
}

// LoadModule will loaded all used module(s).
func (app *App) LoadModule(loader ModuleLoader) {
	if loader == nil {
		return
	}

	// Load module.
	modules := loader.Load()
	if len(modules) == 0 {
		return
	}

	// Interface type.
	moduleInterfaceType := reflect.TypeOf((*Module)(nil)).Elem()
	dependencyInterfaceType := reflect.TypeOf((*Dependency)(nil)).Elem()

	// Scanning modules.
	dependencyMap := make(map[reflect.Type]Dependency)
	for _, module := range modules {
		// Execute all `qore#Module` interface that implemented by module.
		module.HttpRoutes(app)

		moduleVal := reflect.ValueOf(module)
		moduleType := moduleVal.Type()

		// Check is module implements moduleInterfaceType.
		if !moduleType.Implements(moduleInterfaceType) {
			app.Logger().Debug(fmt.Sprintf(
				"Module %s does not implement qore.Module interface", moduleType,
			))
			continue
		}

		// Normalize the module type.
		switch moduleVal.Kind() {
		case reflect.Ptr:
			if moduleVal.Elem().Kind() != reflect.Struct {
				app.Logger().Debug(fmt.Sprintf(
					"Module %s is not a pointer to struct", moduleType,
				))
				continue
			}
			moduleVal = moduleVal.Elem()
			moduleType = moduleType.Elem()
		case reflect.Struct:
			// OK, do nothing
		default:
			app.Logger().Debug(fmt.Sprintf(
				"Module %s does not implements qore.Module interface", moduleType.Elem(),
			))
			continue
		}

		// Get field that implements `qore#Dependency`.
		for i := 0; i < moduleType.NumField(); i++ {
			field := moduleType.Field(i)

			// Must be pointer and implement Dependency interface.
			if field.Type.Kind() != reflect.Ptr {
				continue
			}
			if !field.Type.Implements(dependencyInterfaceType) {
				continue
			}
			// Get actual value from module field.
			depVal := moduleVal.Field(i)
			if !depVal.IsValid() || depVal.IsZero() {
				continue
			}
			// Check is field exported.
			if !field.IsExported() {
				app.Logger().Debug(fmt.Sprintf(
					"Field %s as dependency type in the module %s is private/unexported",
					field.Name, moduleType.Name(),
				))
				continue
			}
			// Cast to Dependency interface.
			depInstance, ok := depVal.Interface().(Dependency)
			if !ok {
				continue
			}
			// Register only once by type.
			if _, exists := dependencyMap[field.Type]; !exists {
				dependencyMap[field.Type] = depInstance
			}
		}
	}
	// Manual memory deallocation.
	defer func() { dependencyMap = nil }()

	// Put dependencies into `app#dependencyRegistry`.
	if len(dependencyMap) == 0 {
		return
	}
	app.dependencyRegistry = make([]Dependency, 0, len(dependencyMap))
	for _, dependency := range dependencyMap {
		app.dependencyRegistry = append(app.dependencyRegistry, dependency)
	}
}

// Start will starting the application services with optional supervisor argument.
// Supervisor is process manager that handle application graceful lifecycle.
// You can create supervisor by yourself using that implement Supervisor interface.
//
// Using default supervisor (SupervisorNon)
//
//	app.Start()
//
// Using gracefully stop & restart supervisor (SupervisorGraceful) backed by overseer
// https://github.com/jpillora/overseer
//
//	app.Start(&qore.SupervisorGraceful{
//		Signals: []os.Signal{
//			syscall.SIGUSR2,
//			syscall.SIGHUP,
//			syscall.SIGTSTP,
//			syscall.SIGINT,
//			os.Interrupt,
//		},
//		BinaryFilePath: "./tmp/main",
//	})
func (app *App) Start(supervisor ...Supervisor) {
	var spv Supervisor = &SupervisorNon{
		Signals: []os.Signal{
			syscall.SIGUSR2,
			syscall.SIGHUP,
			syscall.SIGTSTP,
			syscall.SIGINT,
			os.Interrupt,
		},
	}
	if len(supervisor) > 0 {
		if supervisor[0] != nil {
			spv = supervisor[0]
		}
	}

	// Open dependency.
	if len(app.dependencyRegistry) > 0 {
		sort.Slice(app.dependencyRegistry, func(i, j int) bool {
			return app.dependencyRegistry[i].Priority() < app.dependencyRegistry[j].Priority()
		})
		for _, dependency := range app.dependencyRegistry {
			if err := dependency.Open(); err != nil {
				err = fmt.Errorf("dependency %s failed to open: %w", dependency.Name(), err)
				app.Logger().Error(err.Error())
				continue
			}
			app.Logger().Debug(fmt.Sprintf("dependency %s has been opened successfully", dependency.Name()))
		}
	}

	// Run the application inside supervisor.
	spv.Run(app)

	// Close dependency.
	if len(app.dependencyRegistry) > 0 {
		sort.Slice(app.dependencyRegistry, func(i, j int) bool {
			return app.dependencyRegistry[i].Priority() > app.dependencyRegistry[j].Priority()
		})
		for _, dependency := range app.dependencyRegistry {
			if err := dependency.Close(); err != nil {
				err = fmt.Errorf("dependency %s failed to close: %w", dependency.Name(), err)
				app.Logger().Error(err.Error())
				continue
			}
			app.Logger().Debug(fmt.Sprintf("dependency %s has been closed successfully", dependency.Name()))
		}
	}
}

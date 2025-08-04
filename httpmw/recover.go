package httpmw

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/qoinlyid/qore"
)

// RecoverConfig defines the config for Recover middleware.
type RecoverConfig struct {
	// Size of the stack to be printed.
	// Optional. Default value 4KB.
	StackSize int
	// DisableStackAll disables formatting stack traces of all other goroutines
	// into buffer after the trace for the current goroutine.
	// Optional. Default value false.
	DisableStackAll bool
	// DisablePrintStack disables printing stack trace.
	// Optional. Default value as false.
	DisablePrintStack bool
}

var defaultRecoverConfig = RecoverConfig{StackSize: 4 << 10}

func recoverHandler(next qore.HttpHandler, config RecoverConfig) qore.HttpHandler {
	return func(c qore.HttpContext) (e error) {
		defer func() {
			if r := recover(); r != nil {
				if r == http.ErrAbortHandler {
					panic(r)
				}
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", err)
				}

				var (
					stack  []byte
					length int
				)
				if !config.DisablePrintStack {
					stack = make([]byte, config.StackSize)
					length = runtime.Stack(stack, !config.DisableStackAll)
					stack = stack[:length]
				}
				log.Printf("[PANIC RECOVER] %v %s\n", err, stack)

				if err != nil {
					c.Error(err)
				} else {
					e = err
				}
			}
		}()

		return next(c)
	}
}

// RecoverWithConfig returns a Recover middleware with config.
// See: `Recover()`.
func RecoverWithConfig(config RecoverConfig) qore.HttpMiddleware {
	// Get config or default.
	if config == (RecoverConfig{}) {
		config = defaultRecoverConfig
	}

	// Return qore.HttpHandler.
	return func(next qore.HttpHandler) qore.HttpHandler {
		return recoverHandler(next, config)
	}
}

// Recover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func Recover(next qore.HttpHandler) qore.HttpHandler {
	return RecoverWithConfig(defaultRecoverConfig)(next)
}

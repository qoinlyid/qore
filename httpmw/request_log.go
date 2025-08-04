package httpmw

import (
	"log/slog"
	"time"

	"github.com/qoinlyid/qore"
)

// RequestLogConfig defines the config for RequestLog middleware.
type RequestLogConfig struct {
	// LogLatency instructs logger to record duration it took to execute rest of the handler chain (next(c) call).
	LogLatency bool
	// LogLatencyHuman instructs logger to record duration it took to execute rest of the handler chain (next(c) call) in the human readable format.
	LogLatencyHuman bool
	// LogProtocol instructs logger to extract request protocol (i.e. `HTTP/1.1` or `HTTP/2`)
	LogProtocol bool
	// LogRemoteIP instructs logger to extract request remote IP. See `echo.Context.RealIP()` for implementation details.
	LogRemoteIP bool
	// LogHost instructs logger to extract request host value (i.e. `example.com`)
	LogHost bool
	// LogMethod instructs logger to extract request method value (i.e. `GET` etc)
	LogMethod bool
	// LogURI instructs logger to extract request URI (i.e. `/list?lang=en&page=1`)
	LogURI bool
	// LogRoutePath instructs logger to extract route path part to which request was matched to (i.e. `/user/:id`)
	LogRoutePath bool
	// LogTraceID instructs logger to extract request ID from request `X-Trace-ID` header or response if request did not have value.
	LogTraceID bool
	// LogReferer instructs logger to extract request referer values.
	LogReferer bool
	// LogUserAgent instructs logger to extract request user agent values.
	LogUserAgent bool
	// LogStatus instructs logger to extract response status code. If handler chain returns an echo.HTTPError,
	// the status code is extracted from the echo.HTTPError returned
	LogStatus bool
	// LogHeaders instructs logger to extract given list of headers from request. Note: request can contain more than
	// one header with same value so slice of values is been logger for each given header.
	//
	// Note: header values are converted to canonical form with http.CanonicalHeaderKey as this how request parser converts header
	// names to. For example, the canonical key for "accept-encoding" is "Accept-Encoding".
	LogHeaders []string
}

// DefaultRequestLogConfig is RequestLog default config.
var DefaultRequestLogConfig = &RequestLogConfig{
	LogLatency:      true,
	LogLatencyHuman: true,
	LogRemoteIP:     true,
	LogMethod:       true,
	LogURI:          true,
	LogRoutePath:    true,
	LogTraceID:      true,
	LogUserAgent:    true,
	LogStatus:       true,
}

func requestLogHandler(next qore.HttpHandler, config *RequestLogConfig) qore.HttpHandler {
	if config == nil {
		config = DefaultRequestLogConfig
	}

	return func(c qore.HttpContext) (e error) {
		var (
			reqArgs []any
			resArgs []any
		)
		req := c.Request()
		res := c.Response()
		start := time.Now()
		logger := c.Log()

		// Check config [LogLatency].
		if config.LogLatency {
			latency := time.Since(start)
			reqArgs = append(reqArgs, slog.Int64("latency", latency.Milliseconds()))
			// Check config [LogLatencyHuman].
			if config.LogLatencyHuman {
				reqArgs = append(reqArgs, slog.String("latencyHuman", latency.String()))
			}
		}
		// Check config [LogProtocol].
		if config.LogProtocol {
			reqArgs = append(reqArgs, slog.String("proto", req.Proto))
		}
		// Check config [LogRemoteIP].
		if config.LogRemoteIP {
			reqArgs = append(reqArgs, slog.String("remoteIP", c.RealIP()))
		}
		// Check config [LogHost].
		if config.LogHost {
			reqArgs = append(reqArgs, slog.String("host", req.Host))
		}
		// Check config [LogMethod].
		if config.LogMethod {
			reqArgs = append(reqArgs, slog.String("method", req.Method))
		}
		// Check config [LogURI].
		if config.LogURI {
			reqArgs = append(reqArgs, slog.String("uri", req.RequestURI))
		}
		// Check config [LogRoutePath].
		if config.LogRoutePath {
			reqArgs = append(reqArgs, slog.String("path", c.Path()))
		}
		// Check config [LogTraceID].
		if config.LogTraceID {
			xtraceid := req.Header.Get(qore.HTTP_HEADER_TRACE_ID)
			if xtraceid == "" {
				xtraceid = res.Header().Get(qore.HTTP_HEADER_TRACE_ID)
			}
			logger = logger.With("traceId", xtraceid)
		}
		// Check config [LogReferer].
		if config.LogReferer {
			reqArgs = append(reqArgs, slog.String("referer", req.Referer()))
		}
		// Check config [LogUserAgent].
		if config.LogUserAgent {
			reqArgs = append(reqArgs, slog.String("userAgent", req.UserAgent()))
		}
		// Check config [LogStatus].
		if config.LogStatus {
			resArgs = append(resArgs, slog.Int("status", res.Status))
		}
		// Check config [LogHeaders].
		if len(config.LogHeaders) > 0 {
			var x []any
			for _, key := range config.LogHeaders {
				if val := req.Header.Get(key); val != "" {
					x = append(x, slog.String(key, val))
				}
			}
			reqArgs = append(reqArgs, slog.Group("header", x...))
		}

		logger.Info(
			"RequestLog",
			slog.Group("request", reqArgs...),
			slog.Group("response", resArgs...),
		)
		return next(c)
	}
}

// RequestLogWithConfig returns a RequestLog middleware with config.
// See: `RequestLog()`.
func RequestLogWithConfig(config *RequestLogConfig) qore.HttpMiddleware {
	// Return qore.HttpHandler.
	return func(next qore.HttpHandler) qore.HttpHandler {
		return requestLogHandler(next, config)
	}
}

// RequestLog returns a request logging middleware.
func RequestLog(next qore.HttpHandler) qore.HttpHandler {
	return RequestLogWithConfig(DefaultRequestLogConfig)(next)
}

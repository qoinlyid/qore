package qore

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// HttpValidator is the interface that wrap the Validate function.
type HttpValidator interface {
	// Validate is a function to validate the given interface and return error if any.
	Validate(a any) *HttpValidatorErr
}

// ApiResponseInterface defines the `HttpResponder` interface to creating the instance.
type ApiResponseInterface interface {
	New(c HttpContext) ApiResponse
}

// ApiResponse defines the response wrapper interface for HTTP(s) API.
type ApiResponse interface {
	// Success returns `ApiResponse` with HTTP(s) success status (2xx).
	Success(status HttpResponseSuccess, data ...any) ApiResponse
	// Error returns `ApiResponse` but should be check the error,
	// so it may return client or server error based on given error.
	Error(err error) ApiResponse
	// ClientError returns `ApiResponse` with HTTP(s) client error status (4xx).
	ClientError(status HttpResponseClientError, err error) ApiResponse
	// ServerError returns `ApiResponse` with HTTP(s) server error status (5xx).
	ServerError(status HttpResponseServerError, err error) ApiResponse
	// WithMessage returns `ApiResponse` with response message.
	WithMessage(msg string) ApiResponse
	// WithMessage returns `ApiResponse` with response code.
	WithCode(code any) ApiResponse
	// WithMessage returns `ApiResponse` with additional data in the response.
	WithAdditional(data any) ApiResponse
	// Response write the `ApiResponse`. Returning error or nil.
	Response() error
}

// HttpContext represents the context of the current HTTP(s) request.
// It holds request and response objects, path, path parameters, data and registered handler.
// It is abstraction copy from echo.Context.
type HttpContext interface {
	// Request returns `*http.Request`.
	Request() *http.Request
	// SetRequest sets `*http.Request`.
	SetRequest(r *http.Request)
	// SetResponse sets `*Response`.
	SetResponse(r *echo.Response)
	// Response returns `*Response`.
	Response() *echo.Response
	// IsTLS returns true if HTTP connection is TLS otherwise false.
	IsTLS() bool
	// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
	IsWebSocket() bool
	// Scheme returns the HTTP protocol scheme, `http` or `https`.
	Scheme() string
	// RealIP returns the client's network address based on `X-Forwarded-For`
	// or `X-Real-IP` request header.
	// The behavior can be configured using `Echo#IPExtractor`.
	RealIP() string
	// Path returns the registered path for the handler.
	Path() string
	// SetPath sets the registered path for the handler.
	SetPath(p string)
	// Param returns path parameter by name.
	Param(name string) string
	// ParamNames returns path parameter names.
	ParamNames() []string
	// SetParamNames sets path parameter names.
	SetParamNames(names ...string)
	// ParamValues returns path parameter values.
	ParamValues() []string
	// SetParamValues sets path parameter values.
	SetParamValues(values ...string)
	// QueryParam returns the query param for the provided name.
	QueryParam(name string) string
	// QueryParams returns the query parameters as `url.Values`.
	QueryParams() url.Values
	// QueryString returns the URL query string.
	QueryString() string
	// FormValue returns the form field value for the provided name.
	FormValue(name string) string
	// FormParams returns the form parameters as `url.Values`.
	FormParams() (url.Values, error)
	// FormFile returns the multipart form file for the provided name.
	FormFile(name string) (*multipart.FileHeader, error)
	// MultipartForm returns the multipart form.
	MultipartForm() (*multipart.Form, error)
	// Cookie returns the named cookie provided in the request.
	Cookie(name string) (*http.Cookie, error)
	// SetCookie adds a `Set-Cookie` header in HTTP response.
	SetCookie(cookie *http.Cookie)
	// Cookies returns the HTTP cookies sent with the request.
	Cookies() []*http.Cookie
	// Get retrieves data from the context.
	Get(key string) any
	// Set saves data in the context.
	Set(key string, val any)
	// Bind binds path params, query params and the request body into provided type `i`. The default binder
	// binds body based on Content-Type header.
	Bind(i any) error
	// Validate validates provided `i`. It is usually called after `Context#Bind()`.
	// Validator must be registered using `Echo#Validator`.
	Validate(i any) error
	// Render renders a template with data and sends a text/html response with status
	// code. Renderer must be registered using `Echo.Renderer`.
	Render(code int, name string, data any) error
	// HTML sends an HTTP response with status code.
	HTML(code int, html string) error
	// HTMLBlob sends an HTTP blob response with status code.
	HTMLBlob(code int, b []byte) error
	// String sends a string response with status code.
	String(code int, s string) error
	// JSON sends a JSON response with status code.
	JSON(code int, i any) error
	// JSONPretty sends a pretty-print JSON with status code.
	JSONPretty(code int, i any, indent string) error
	// JSONBlob sends a JSON blob response with status code.
	JSONBlob(code int, b []byte) error
	// JSONP sends a JSONP response with status code. It uses `callback` to construct
	// the JSONP payload.
	JSONP(code int, callback string, i any) error
	// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
	// to construct the JSONP payload.
	JSONPBlob(code int, callback string, b []byte) error
	// XML sends an XML response with status code.
	XML(code int, i any) error
	// XMLPretty sends a pretty-print XML with status code.
	XMLPretty(code int, i any, indent string) error
	// XMLBlob sends an XML blob response with status code.
	XMLBlob(code int, b []byte) error
	// Blob sends a blob response with status code and content type.
	Blob(code int, contentType string, b []byte) error
	// Stream sends a streaming response with status code and content type.
	Stream(code int, contentType string, r io.Reader) error
	// File sends a response with the content of the file.
	File(file string) error
	// Attachment sends a response as attachment, prompting client to save the
	// file.
	Attachment(file string, name string) error
	// Inline sends a response as inline, opening the file in the browser.
	Inline(file string, name string) error
	// NoContent sends a response with no body and a status code.
	NoContent(code int) error
	// Redirect redirects the request to a provided URL with status code.
	Redirect(code int, url string) error
	// Error invokes the registered global HTTP error handler. Generally used by middleware.
	// A side-effect of calling global error handler is that now Response has been committed (sent to the client) and
	// middlewares up in chain can not change Response status code or Response body anymore.
	//
	// Avoid using this method in handlers as no middleware will be able to effectively handle errors after that.
	Error(err error)
	// Reset resets the context after request completes. It must be called along
	// with `Echo#AcquireContext()` and `Echo#ReleaseContext()`.
	// See `Echo#ServeHTTP()`
	Reset(r *http.Request, w http.ResponseWriter)
	// ValidateRequest validates provided request payload `i`. It is usually called after `Context#Bind()`.
	// Validator must be registered using `App#SetHttpValidator()`.
	ValidateRequest(i any) *HttpValidatorErr
	// Log.
	TraceID() (val string)
	// Log.
	Log() *logger
	// Api return HTTP(s) API responder.
	Api() ApiResponse
}

// HttpHandler defines a function to serve HTTP(s) requests.
type HttpHandler func(c HttpContext) error

// HttpMiddleware defines a function to process middleware.
type HttpMiddleware func(next HttpHandler) HttpHandler

// HttpHandlerable defines HTTP(s) that can be used by user's .
type HttpHandlerable HttpHandler

// HttpHandlerableWithPayload same like `HttpHandlerable` but need & use request payload.
type HttpHandlerableWithPayload[T HttpRequestPayload] func(c HttpContext, p T) error

// HttpHandlerChain defines HTTP(s) handler wrapper.
type HttpHandlerChain[T HttpRequestPayload] interface {
	// HandlerWrapper is wrapper function for HTTP(s) handler.
	HandlerWrapper(c HttpContext) error
}

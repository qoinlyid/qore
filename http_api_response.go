package qore

import (
	"context"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
)

// ApiResponseDefault defines default API response object.
type ApiResponseDefault struct {
	Success bool   `json:"success" xml:"success"`
	Code    string `json:"code" xml:"code"`
	Data    any    `json:"data" xml:"data"`
	Error   any    `json:"error" xml:"error"`
}

type apiResponseInterfaceImpl struct{}

type apiResponseImpl struct {
	ctx    HttpContext
	status int
	object *ApiResponseDefault
}

// Compile time check `apiResponseInterfaceImpl implements `ApiResponseInterface`.
var _ ApiResponseInterface = apiResponseInterfaceImpl{}

// Compile time check `apiResponseImpl implements `ApiResponse`.
var _ ApiResponse = (*apiResponseImpl)(nil)

// New creates `ApiResponse`.
func (i apiResponseInterfaceImpl) New(c HttpContext) ApiResponse {
	return &apiResponseImpl{ctx: c, object: new(ApiResponseDefault)}
}

// Success returns `ApiResponse` with HTTP(s) success status (2xx).
func (r *apiResponseImpl) Success(status HttpResponseSuccess, data ...any) ApiResponse {
	r.status = int(status)
	r.object.Success = true
	r.object.Code = fmt.Sprintf("%d", r.status)
	if len(data) == 1 {
		r.object.Data = data[0]
	} else {
		r.object.Data = data
	}
	return r
}

// Error returns `ApiResponse` but should be check the error,
// so it may return client or server error based on given error.
func (r *apiResponseImpl) Error(err error) ApiResponse {
	if err == nil {
		return r.ServerError(HttpStatusNotImplemented, errors.New("failed: argument error is null"))
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return r.ServerError(HttpStatusGatewayTimeout, errors.New("failed: Request timeout"))
	case errors.Is(err, context.Canceled):
		return r.ServerError(HttpStatusServiceUnavailable, errors.New("failed: Request canceled"))
	default:
		return r.ServerError(HttpStatusInternalServerError, err)
	}
}

// ClientError returns `ApiResponse` with HTTP(s) client error status (4xx).
func (r *apiResponseImpl) ClientError(status HttpResponseClientError, err error) ApiResponse {
	r.status = int(status)
	r.object.Success = false
	r.object.Code = fmt.Sprintf("%d", r.status)
	if err != nil {
		r.object.Error = err.Error()
	}
	return r
}

// ServerError returns `ApiResponse` with HTTP(s) server error status (5xx).
func (r *apiResponseImpl) ServerError(status HttpResponseServerError, err error) ApiResponse {
	r.status = int(status)
	r.object.Success = false
	r.object.Code = fmt.Sprintf("%d", r.status)
	if err != nil {
		r.object.Error = err.Error()
	}
	return r
}

// WithMessage returns `ApiResponse` with response message.
func (r *apiResponseImpl) WithMessage(msg string) ApiResponse { return r }

// WithMessage returns `ApiResponse` with response code.
func (r *apiResponseImpl) WithCode(code any) ApiResponse {
	if code != nil {
		r.object.Code = fmt.Sprintf("%v", code)
	}
	return r
}

// WithMessage returns `ApiResponse` with additional data in the response.
func (r *apiResponseImpl) WithAdditional(data any) ApiResponse { return r }

// Response write the `ApiResponse`. Returning error or nil.
func (r *apiResponseImpl) Response() error {
	accept := r.ctx.Request().Header.Get("Accept")
	switch accept {
	case echo.MIMETextXML, echo.MIMEApplicationXML, echo.MIMEApplicationXMLCharsetUTF8:
		return r.ctx.XML(r.status, r.object)
	default:
		return r.ctx.JSON(r.status, r.object)
	}
}

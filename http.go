package qore

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type httpServer struct {
	core            *echo.Echo
	autoTLS         bool
	certPath        string
	keyPath         string
	shutdownTimeout time.Duration
	validator       HttpValidator
	iApiResponse    ApiResponseInterface
}

func newHttpServer() *httpServer {
	core := echo.New()
	core.HideBanner = true
	core.HidePort = true
	return &httpServer{
		core:         core,
		validator:    httpValidatorDefault(),
		iApiResponse: apiResponseInterfaceImpl{},
	}
}

func (s *httpServer) start(lfn func() (listener net.Listener, err error), logger *logger) {
	logger = logger.With(slog.String("scope", "http(s) server"))
	if s.core == nil {
		logger.Warn("http(s) server doe not initiated yet")
		return
	}
	listener, err := lfn()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	go func() {
		s.core.Listener = listener
		switch {
		case s.autoTLS:
			err = s.core.StartAutoTLS("")
		case !ValidationIsEmpty(s.certPath) && !ValidationIsEmpty(s.keyPath):
			err = s.core.StartTLS("", s.certPath, s.keyPath)
		default:
			err = s.core.Start("")
		}
	}()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err.Error())
			return
		}
	}
	logger.Debug("HTTP(S) server running...")
}

func (s *httpServer) stop(logger *logger) {
	if s.core == nil {
		return
	}
	logger = logger.With(slog.String("scope", "http(s) server"))

	// Shutdown with context timeout.
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	err := s.core.Shutdown(ctx)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err.Error())
			return
		}
	}
	logger.Debug("HTTP(S) server was stoped")
}

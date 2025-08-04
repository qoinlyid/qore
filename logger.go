package qore

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

type logger struct {
	base      *slog.Logger
	addSource bool
}

func setupLogger(config *Config) *logger {
	// Logging level.
	level := slog.LevelError
	switch config.LogLevel {
	case LOG_DEBUG:
		level = slog.LevelDebug
	case LOG_INFO:
		level = slog.LevelInfo
	case LOG_WARN:
		level = slog.LevelWarn
	}

	// Set logger handler.
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: config.LogShowSource,
		Level:     level,
	})
	if config.LogJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.LogShowSource,
			Level:     level,
		})
	}

	return &logger{
		base:      slog.New(handler),
		addSource: config.LogShowSource,
	}
}

func (l *logger) clone() *logger {
	c := *l
	return &c
}

func (l *logger) record(level slog.Level, msg string, args ...any) {
	ctx := context.Background()
	if !l.base.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if l.addSource {
		x, _, _, ok := runtime.Caller(2)
		if ok {
			pc = x
		}
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	l.base.Handler().Handle(ctx, r)
}

func (l *logger) Group(name string) *logger {
	if ValidationIsEmpty(name) {
		return l
	}
	c := l.clone()
	c.base = l.base.WithGroup(name)
	return c
}

func (l *logger) With(args ...any) *logger {
	if len(args) == 0 {
		return l
	}
	c := l.clone()
	c.base = l.base.With(args...)
	return c
}

func (l *logger) Debug(msg string, args ...any) { l.record(slog.LevelDebug, msg, args...) }
func (l *logger) Info(msg string, args ...any)  { l.record(slog.LevelInfo, msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { l.record(slog.LevelWarn, msg, args...) }
func (l *logger) Error(msg string, args ...any) { l.record(slog.LevelError, msg, args...) }

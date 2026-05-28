package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type Options struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output io.Writer
}

var (
	mu            sync.RWMutex
	defaultLogger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
)

func Init(opts Options) {
	mu.Lock()
	defer mu.Unlock()
	defaultLogger = newLogger(opts)
}

func Reconfigure(opts Options) {
	Init(opts)
}

func Default() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLogger
}

func WithContext(ctx context.Context) *Logger {
	return &Logger{
		ctx:    ctx,
		logger: loggerFromContext(ctx),
	}
}

type Logger struct {
	ctx    context.Context
	logger *slog.Logger
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.DebugContext(l.ctx, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.InfoContext(l.ctx, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.WarnContext(l.ctx, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.ErrorContext(l.ctx, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.logger.ErrorContext(l.ctx, msg, args...)
	os.Exit(1)
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

func newLogger(opts Options) *slog.Logger {
	level := parseLevel(opts.Level)
	out := opts.Output
	if out == nil {
		out = os.Stderr
	}

	handlerOpts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "text":
		handler = slog.NewTextHandler(out, handlerOpts)
	default:
		handler = slog.NewJSONHandler(out, handlerOpts)
	}

	return slog.New(handler)
}

func parseLevel(level string) slog.Leveler {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func loggerFromContext(ctx context.Context) *slog.Logger {
	l := defaultLogger
	if ctx == nil {
		return l
	}

	attrs := contextAttrs(ctx)
	if len(attrs) == 0 {
		return l
	}
	return l.With(attrs...)
}

func contextAttrs(ctx context.Context) []any {
	var attrs []any
	if requestID, ok := RequestIDFromContext(ctx); ok {
		attrs = append(attrs, "request_id", requestID)
	}
	return attrs
}

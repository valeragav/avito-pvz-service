package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type loggerKey struct{}

type Logger struct {
	logger *slog.Logger
}

var globalLog *Logger

func New(appName, env, level string) *Logger {
	cnfLog := ConfigureLogger(appName, env, level)
	return NewLogger(cnfLog)
}

func NewLogger(logger *slog.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func MustSetGlobal(l *Logger) {
	if l == nil {
		panic("logger is nil")
	}
	globalLog = l
}

func ConfigureLogger(appName, env, logLevel string) *slog.Logger {
	level := parseLogLevel(logLevel)

	opts := slog.HandlerOptions{
		Level:     level,
		AddSource: env == "dev",
	}

	handler := &ctxHandler{
		Handler: slog.NewTextHandler(os.Stdout, &opts),
	}

	return slog.New(handler).With(
		slog.String("app", appName),
		slog.String("env", env),
	)
}

func GetLogger() *Logger {
	return globalLog
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, args ...any) {
	l.logger.DebugContext(ctx, msg, args...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, args ...any) {
	l.logger.InfoContext(ctx, msg, args...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, args ...any) {
	l.logger.WarnContext(ctx, msg, args...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, args ...any) {
	l.logger.ErrorContext(ctx, msg, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		logger: l.logger.With(args...),
	}
}

func (l *Logger) IntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func FromCtx(ctx context.Context) *Logger {
	if l, ok := ctx.Value(loggerKey{}).(*Logger); ok {
		return l
	}

	if globalLog != nil {
		return globalLog
	}

	return &Logger{
		logger: slog.Default(),
	}
}

func With(ctx context.Context, keyValues ...any) *Logger {
	return FromCtx(ctx).With(keyValues...)
}

func Info(msg string, args ...any) {
	logWithError(context.Background(), slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	logWithError(context.Background(), slog.LevelWarn, msg, args...)
}

func Debug(msg string, args ...any) {
	logWithError(context.Background(), slog.LevelDebug, msg, args...)
}

func Error(msg string, args ...any) {
	logWithError(context.Background(), slog.LevelError, msg, args...)
}

func Fatal(msg string, args ...any) {
	logWithError(context.Background(), slog.LevelError, msg, args...)
	os.Exit(1)
}

func WithCtx(ctx context.Context, keyValues ...any) context.Context {
	log := FromCtx(ctx).With(keyValues...)
	return context.WithValue(ctx, loggerKey{}, log)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	logWithError(ctx, slog.LevelInfo, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	logWithError(ctx, slog.LevelWarn, msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	logWithError(ctx, slog.LevelDebug, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logWithError(ctx, slog.LevelError, msg, args...)
}

func FatalCtx(ctx context.Context, msg string, args ...any) {
	logWithError(ctx, slog.LevelError, msg, args...)
	os.Exit(1)
}

func logWithError(ctx context.Context, level slog.Level, msg string, args ...any) {
	log := FromCtx(ctx)

	switch level {
	case slog.LevelDebug:
		log.logger.DebugContext(ctx, msg, args...)
	case slog.LevelInfo:
		log.logger.InfoContext(ctx, msg, args...)
	case slog.LevelWarn:
		log.logger.WarnContext(ctx, msg, args...)
	case slog.LevelError:
		log.logger.ErrorContext(ctx, msg, args...)
	default:
		log.logger.InfoContext(ctx, msg, args...)
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

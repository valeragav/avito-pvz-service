package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	logInstance *slog.Logger
	once        sync.Once
)

func GetLogger() *slog.Logger {
	once.Do(func() {
		if logInstance == nil {
			logInstance = InitLogger("unknown", "dev", "info")
		}
	})
	return logInstance
}

// InitLogger создаёт и настраивает глобальный логгер
func InitLogger(appName, env, logLevel string) *slog.Logger {
	level := parseLogLevel(logLevel)

	opts := slog.HandlerOptions{
		Level:     level,
		AddSource: env == "dev",
	}

	handler := &ctxHandler{
		Handler: slog.NewTextHandler(os.Stdout, &opts),
	}

	logger := slog.New(handler).With(
		slog.String("app", appName),
		slog.String("env", env),
	)

	logInstance = logger
	return logInstance
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

func Fatal(msg string, args ...any) {
	// TODO: добавить уровень Fatal
	LogCtx(context.Background(), slog.LevelError, msg, args...)
	os.Exit(1)
}

func Debug(msg string, args ...any) {
	LogCtx(context.Background(), slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	LogCtx(context.Background(), slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	LogCtx(context.Background(), slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	LogCtx(context.Background(), slog.LevelError, msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	LogCtx(ctx, slog.LevelDebug, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	LogCtx(ctx, slog.LevelInfo, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	LogCtx(ctx, slog.LevelWarn, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	LogCtx(ctx, slog.LevelError, msg, args...)
}

func LogCtx(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger := GetLogger()

	switch level {
	case slog.LevelDebug:
		logger.DebugContext(ctx, msg, args...)
	case slog.LevelInfo:
		logger.InfoContext(ctx, msg, args...)
	case slog.LevelWarn:
		logger.WarnContext(ctx, msg, args...)
	case slog.LevelError:
		logger.ErrorContext(ctx, msg, args...)
	default:
		logger.InfoContext(ctx, msg, args...)
	}
}

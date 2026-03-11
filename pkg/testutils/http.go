package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"sync"

	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

var initLoggerOnce sync.Once

func InitTestLogger() {
	initLoggerOnce.Do(func() {
		opts := slog.HandlerOptions{
			Level: slog.LevelError,
		}
		handler := slog.NewTextHandler(io.Discard, &opts)
		logTest := logger.NewLogger(slog.New(handler))
		logger.MustSetGlobal(logTest)
	})
}

func MakeRequestBody(body any) (*bytes.Reader, error) {
	if str, ok := body.(string); ok {
		return bytes.NewReader([]byte(str)), nil
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

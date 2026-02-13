package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/VaLeraGav/avito-pvz-service/pkg/logger"
)

func InitTestLogger() {
	opts := slog.HandlerOptions{
		Level: slog.LevelError,
	}

	handler := slog.NewTextHandler(io.Discard, &opts)
	logTest := logger.NewLogger(slog.New(handler))
	logger.MustSetGlobal(logTest)
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

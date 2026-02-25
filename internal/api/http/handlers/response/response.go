package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Empty struct{}

type Error struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, ctx context.Context, status int, errorMsg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Error{
		Message: errorMsg,
	}

	if err != nil {
		// NOTE: remove it for the user, and leave it for the moderator
		response.Details = err.Error()
	}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(ctx, "failed to encode error response", "error", err)
	}
}

func WriteJSON(w http.ResponseWriter, ctx context.Context, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "internal Server Error", http.StatusInternalServerError)
		}
	}
}

func WriteString(w http.ResponseWriter, ctx context.Context, status int, data string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	if _, err := w.Write([]byte(data)); err != nil {
		http.Error(w, "internal Server Error", http.StatusInternalServerError)
	}
}

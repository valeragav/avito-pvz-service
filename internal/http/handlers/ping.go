package handlers

import (
	"net/http"

	"github.com/VaLeraGav/avito-pvz-service/internal/http/handlers/response"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))

	if err != nil {
		response.WriteError(w, r.Context(), http.StatusBadRequest, "internal server error", err)
		return
	}
}

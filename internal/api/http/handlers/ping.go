package handlers

import (
	"net/http"

	"github.com/valeragav/avito-pvz-service/internal/api/http/handlers/response"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("pong"))

	if err != nil {
		response.WriteError(w, r.Context(), http.StatusBadRequest, "internal server error", err)
		return
	}
}

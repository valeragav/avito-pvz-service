package cors

import (
	"net/http"

	"github.com/go-chi/cors"
)

var corsCfg = cors.Options{
	AllowedOrigins: []string{
		"http://localhost:8107",
	},
	AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	AllowedHeaders: []string{
		"Content-Type",
		"Accept",
		"Authorization",
		"X-CSRF-Token",
		"X-Request-ID",
		"Device-Uid",
	},
	ExposedHeaders: []string{
		"Content-Type",
		"Link",
		"X-Request-ID",
		"Device-Uid",
	},
	AllowCredentials: true,
	MaxAge:           300,
}

func Handler(next http.Handler) http.Handler {
	corsHandler := cors.New(corsCfg)
	return corsHandler.Handler(next)
}

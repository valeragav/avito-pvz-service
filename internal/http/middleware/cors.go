package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

func corsCfg(allowedOrigins []string) cors.Options {
	return cors.Options{
		AllowedOrigins: allowedOrigins,
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
}

func Cors(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		corsHandler := cors.New(corsCfg(allowedOrigins))
		return corsHandler.Handler(next)
	}
}

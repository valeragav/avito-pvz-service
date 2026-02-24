package middleware

import (
	"net/http"
)

// Concurrency возвращает middleware, ограничивающий число одновременно обрабатываемых запросов.
// При достижении лимита новые запросы немедленно получают 503 Service Unavailable,
// не занимая горутину в ожидании. Это защищает пул соединений к БД от перегрузки:
// при 1000 RPS × SLI 100ms в параллели находится ≤ 100 запросов — semaphore на 500
// даёт двукратный запас и при этом не допускает лавинообразного роста соединений.
func Concurrency(maxRequests int) func(http.Handler) http.Handler {
	if maxRequests <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	sem := make(chan struct{}, maxRequests)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.ServeHTTP(w, r)
			default:
				http.Error(w, "server is overloaded, please retry", http.StatusServiceUnavailable)
			}
		})
	}
}

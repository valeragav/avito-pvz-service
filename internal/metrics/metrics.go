package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func Init() {
	prometheus.MustRegister(
		createdPvz,
		createdProducts,
		createdReceptions,
		httpRequestsTotal,
		httpRequestDuration,
		httpResponsesTotal,
	)
}

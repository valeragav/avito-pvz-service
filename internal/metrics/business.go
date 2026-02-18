package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: сделать отдельную структуру, чтобы тесты или при нагрузочном тестировании не портили метрику

var (
	createdPVZ = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_pvz_total",
			Help: "Total number of created Pvz.",
		},
	)

	createdProducts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_products_total",
			Help: "Total number of created products.",
		},
	)

	createdReceptions = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "created_receptions_total",
			Help: "Total number of created receptions.",
		},
	)
)

func CreatedPVZInc() {
	createdPVZ.Inc()
}

func CreatedProductsInc() {
	createdProducts.Inc()
}

func CreatedReceptionsInc() {
	createdReceptions.Inc()
}

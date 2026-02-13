package metrics

import (
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	createdPvz = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "created_pvz_total",
			Help: "Total number of created Pvz.",
		},
		[]string{"city"},
	)

	createdProducts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "created_products_total",
			Help: "Total number of created products.",
		},
		[]string{"pvz_id"},
	)

	createdReceptions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "created_receptions_total",
			Help: "Total number of created receptions.",
		},
		[]string{"pvz_id"},
	)
)

func CreatedPvzInc(city string) {
	createdPvz.WithLabelValues(city).Inc()
}

func CreatedProductsInc(pvzId uuid.UUID) {
	createdProducts.WithLabelValues(pvzId.String()).Inc()
}

func CreatedReceptionsInc(pvzId uuid.UUID) {
	createdReceptions.WithLabelValues(pvzId.String()).Inc()
}

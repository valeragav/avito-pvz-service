package pvz

import (
	"context"

	"github.com/VaLeraGav/avito-pvz-service/internal/service/pvz"
)

//go:generate mockgen -source=contract.go -destination=./mocks/service_mock.go -package=mocks
type productsService interface {
	Create(ctx context.Context, createIn pvz.CreateIn) (*pvz.CreateOut, error)
	List(ctx context.Context, pvzListParams *pvz.PvzListParams) (*pvz.PvzListResponse, error)
}

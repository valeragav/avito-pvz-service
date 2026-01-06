package container

import (
	"log/slog"

	"github.com/VaLeraGav/avito-pvz-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DIContainer struct {
}

// аналог wire.go
func Init(cfg *config.Config, lg *slog.Logger, connPostgres *pgxpool.Pool) *DIContainer {

	return &DIContainer{}
}

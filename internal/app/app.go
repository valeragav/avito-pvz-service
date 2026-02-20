package app

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infra/postgres"
	"github.com/valeragav/avito-pvz-service/internal/security"
	"github.com/valeragav/avito-pvz-service/internal/usecase/auth"
	"github.com/valeragav/avito-pvz-service/internal/usecase/product"
	"github.com/valeragav/avito-pvz-service/internal/usecase/pvz"
	"github.com/valeragav/avito-pvz-service/internal/usecase/reception"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type App struct {
	AuthUseCase      *auth.AuthUseCase
	PVZUseCase       *pvz.PVZUseCase
	ReceptionUseCase *reception.ReceptionUseCase
	ProductUseCase   *product.ProductUseCase

	Validator  *validation.Validator
	JwtService *security.JwtService
}

func New(cfg *config.Config, lg *logger.Logger, db *pgxpool.Pool) (*App, error) {
	// repos
	userRepo := postgres.NewUserRepository(db)
	pvzRepo := postgres.NewPVZRepository(db)
	cityRepo := postgres.NewCityRepository(db)
	receptionRepo := postgres.NewReceptionRepository(db)
	statusRepo := postgres.NewReceptionStatusRepository(db)
	productRepo := postgres.NewProductRepository(db)
	productTypeRepo := postgres.NewProductTypeRepository(db)

	// services
	jwtService, err := security.New(
		cfg.Jwt.RSAPrivateFile,
		cfg.Jwt.RSAPublicFile,
		cfg.Jwt.Iss,
		cfg.Jwt.AccessLifeTime,
	)
	if err != nil {
		return nil, err
	}

	validator := validation.New()

	// usecases
	authUC := auth.New(jwtService, userRepo)
	pvzUC := pvz.New(pvzRepo, cityRepo, receptionRepo, productRepo)
	receptionUC := reception.New(receptionRepo, statusRepo, pvzRepo)
	productUC := product.New(productRepo, receptionRepo, productTypeRepo, pvzRepo)

	return &App{
		AuthUseCase:      authUC,
		PVZUseCase:       pvzUC,
		ReceptionUseCase: receptionUC,
		ProductUseCase:   productUC,

		Validator:  validator,
		JwtService: jwtService,
	}, nil
}

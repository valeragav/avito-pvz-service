package container

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
	"github.com/valeragav/avito-pvz-service/internal/security"
	"github.com/valeragav/avito-pvz-service/internal/usecase/auth"
	"github.com/valeragav/avito-pvz-service/internal/usecase/product"
	"github.com/valeragav/avito-pvz-service/internal/usecase/pvz"
	"github.com/valeragav/avito-pvz-service/internal/usecase/reception"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type DIContainer struct {
	cfg          *config.Config
	lg           *logger.Logger
	connPostgres *pgxpool.Pool

	UserRepo         *repo.UserRepository
	PVZRepo          *repo.PVZRepository
	CitiesRepo       *repo.CityRepository
	ReceptionsRepo   *repo.ReceptionRepository
	StatusesRepo     *repo.ReceptionStatusRepository
	ProductsRepo     *repo.ProductRepository
	ProductTypesRepo *repo.ProductTypeRepository

	JwtService *security.JwtService
	Validator  *validation.Validator

	AuthUseCase      *auth.AuthUseCase
	PVZUseCase       *pvz.PVZUseCase
	ReceptionUseCase *reception.ReceptionUseCase
	ProductUseCase   *product.ProductUseCase
}

// аналог wire.go

func New(cfg *config.Config, lg *logger.Logger, connPostgres *pgxpool.Pool) *DIContainer {
	return &DIContainer{
		cfg:          cfg,
		lg:           lg,
		connPostgres: connPostgres,
	}
}

func (c *DIContainer) Init() error {
	c.UserRepo = repo.NewUserRepository(c.connPostgres)
	c.PVZRepo = repo.NewPVZRepository(c.connPostgres)
	c.CitiesRepo = repo.NewCityRepository(c.connPostgres)
	c.ReceptionsRepo = repo.NewReceptionRepository(c.connPostgres)
	c.StatusesRepo = repo.NewReceptionStatusRepository(c.connPostgres)
	c.ProductsRepo = repo.NewProductRepository(c.connPostgres)
	c.ProductTypesRepo = repo.NewProductTypeRepository(c.connPostgres)

	if _, err := c.GetJwtTokenService(); err != nil {
		return err
	}

	c.GetValidation()

	c.GetAuthUseCase()
	c.GetPVZUseCase()
	c.GetReceptionUseCase()
	c.GetProductUseCase()

	return nil
}

func (c *DIContainer) GetJwtTokenService() (*security.JwtService, error) {
	if c.JwtService != nil {
		return c.JwtService, nil
	}

	jwtUseCase, err := security.New(
		c.cfg.Jwt.RSAPrivateFile,
		c.cfg.Jwt.RSAPublicFile,
		c.cfg.Jwt.Iss,
		c.cfg.Jwt.AccessLifeTime,
	)
	if err != nil {
		return nil, err
	}

	c.JwtService = jwtUseCase
	return c.JwtService, nil
}

func (c *DIContainer) GetValidation() *validation.Validator {
	if c.Validator == nil {
		c.Validator = validation.New()
	}

	return c.Validator
}

func (c *DIContainer) GetAuthUseCase() *auth.AuthUseCase {
	if c.AuthUseCase == nil {
		c.AuthUseCase = auth.New(c.JwtService, c.UserRepo)
	}

	return c.AuthUseCase
}

func (c *DIContainer) GetPVZUseCase() *pvz.PVZUseCase {
	if c.PVZUseCase == nil {
		c.PVZUseCase = pvz.New(c.PVZRepo, c.CitiesRepo, c.ReceptionsRepo, c.ProductsRepo)
	}

	return c.PVZUseCase
}

func (c *DIContainer) GetReceptionUseCase() *reception.ReceptionUseCase {
	if c.ReceptionUseCase == nil {
		c.ReceptionUseCase = reception.New(c.ReceptionsRepo, c.StatusesRepo, c.PVZRepo)
	}

	return c.ReceptionUseCase
}
func (c *DIContainer) GetProductUseCase() *product.ProductUseCase {
	if c.ProductUseCase == nil {
		c.ProductUseCase = product.New(c.ProductsRepo, c.ReceptionsRepo, c.ProductTypesRepo, c.PVZRepo)
	}

	return c.ProductUseCase
}

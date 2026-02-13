package container

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valeragav/avito-pvz-service/internal/config"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/cities"
	productTypesRepo "github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/product_types"
	productsRepo "github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/products"
	pvzRepo "github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/pvz"
	receptionsRepo "github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/receptions"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/statuses"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/user"
	"github.com/valeragav/avito-pvz-service/internal/security"
	"github.com/valeragav/avito-pvz-service/internal/service/auth"
	"github.com/valeragav/avito-pvz-service/internal/service/products"
	"github.com/valeragav/avito-pvz-service/internal/service/pvz"
	"github.com/valeragav/avito-pvz-service/internal/service/receptions"
	"github.com/valeragav/avito-pvz-service/internal/validation"
	"github.com/valeragav/avito-pvz-service/pkg/logger"
)

type DIContainer struct {
	cfg          *config.Config
	lg           *logger.Logger
	connPostgres *pgxpool.Pool

	UserRepo         *user.Repository
	PvzRepo          *pvzRepo.Repository
	CitiesRepo       *cities.Repository
	ReceptionsRepo   *receptionsRepo.Repository
	StatusesRepo     *statuses.Repository
	ProductsRepo     *productsRepo.Repository
	ProductTypesRepo *productTypesRepo.Repository

	JwtService *security.JwtService
	Validator  *validation.Validator

	AuthService       *auth.AuthService
	PvzService        *pvz.PvzService
	ReceptionsService *receptions.ReceptionsService
	ProductsService   *products.ProductsService
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
	c.UserRepo = user.New(c.connPostgres)
	c.PvzRepo = pvzRepo.New(c.connPostgres)
	c.CitiesRepo = cities.New(c.connPostgres)
	c.ReceptionsRepo = receptionsRepo.New(c.connPostgres)
	c.StatusesRepo = statuses.New(c.connPostgres)
	c.ProductsRepo = productsRepo.New(c.connPostgres)
	c.ProductTypesRepo = productTypesRepo.New(c.connPostgres)

	if _, err := c.GetJwtTokenService(); err != nil {
		return err
	}

	c.GetValidation()

	c.GetAuthService()
	c.GetPvzService()
	c.GetReceptionsService()
	c.GetProductsService()

	return nil
}

func (c *DIContainer) GetJwtTokenService() (*security.JwtService, error) {
	if c.JwtService != nil {
		return c.JwtService, nil
	}

	jwtService, err := security.New(
		c.cfg.Jwt.RSAPrivateFile,
		c.cfg.Jwt.RSAPublicFile,
		c.cfg.Jwt.Iss,
		c.cfg.Jwt.AccessLifeTime,
	)
	if err != nil {
		return nil, err
	}

	c.JwtService = jwtService
	return c.JwtService, nil
}

func (c *DIContainer) GetValidation() *validation.Validator {
	if c.Validator == nil {
		c.Validator = validation.New()
	}

	return c.Validator
}

func (c *DIContainer) GetAuthService() *auth.AuthService {
	if c.AuthService == nil {
		c.AuthService = auth.New(c.JwtService, c.UserRepo)
	}

	return c.AuthService
}

func (c *DIContainer) GetPvzService() *pvz.PvzService {
	if c.PvzService == nil {
		c.PvzService = pvz.New(c.PvzRepo, c.CitiesRepo, c.ReceptionsRepo, c.ProductsRepo)
	}

	return c.PvzService
}

func (c *DIContainer) GetReceptionsService() *receptions.ReceptionsService {
	if c.ReceptionsService == nil {
		c.ReceptionsService = receptions.New(c.ReceptionsRepo, c.StatusesRepo)
	}

	return c.ReceptionsService
}
func (c *DIContainer) GetProductsService() *products.ProductsService {
	if c.ProductsService == nil {
		c.ProductsService = products.New(c.ProductsRepo, c.ReceptionsRepo, c.ProductTypesRepo, c.PvzRepo)
	}

	return c.ProductsService
}

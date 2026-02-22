package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/dto"
	"github.com/valeragav/avito-pvz-service/internal/infra"
)

//go:generate ${LOCAL_BIN}/mockgen -source=product.go -destination=./mocks/product_mock.go -package=mocks
type productRepo interface {
	Create(ctx context.Context, product domain.Product) (*domain.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	GetLastProductInReception(ctx context.Context, receptionID uuid.UUID) (*domain.Product, error)
}

type receptionRepo interface {
	FindByStatus(ctx context.Context, statusName domain.ReceptionStatusCode, filter domain.Reception) (*domain.Reception, error)
}

type productTypeRepo interface {
	Get(ctx context.Context, filter domain.ProductType) (*domain.ProductType, error)
}

type pvzRepo interface {
	Get(ctx context.Context, filter domain.PVZ) (*domain.PVZ, error)
}

type ProductUseCase struct {
	productRepo     productRepo
	receptionRepo   receptionRepo
	productTypeRepo productTypeRepo
	pvzRepo         pvzRepo
}

func New(productRepo productRepo, receptionRepo receptionRepo, productTypeRepo productTypeRepo, pvzRepo pvzRepo) *ProductUseCase {
	return &ProductUseCase{
		productRepo,
		receptionRepo,
		productTypeRepo,
		pvzRepo,
	}
}

func (s *ProductUseCase) Create(ctx context.Context, createIn dto.ProductCreate) (*domain.Product, error) {
	const op = "products.Create"

	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: createIn.PvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrNoReceptionIsCurrentlyInProgress
		}
		return nil, fmt.Errorf("%s: failed to find in progress reception: %w", op, err)
	}

	productType, err := s.productTypeRepo.Get(ctx, domain.ProductType{Name: createIn.TypeName})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find product type '%s': %w", op, createIn.TypeName, err)
	}

	product, err := s.productRepo.Create(ctx, domain.Product{
		DateTime:    time.Now(),
		TypeID:      productType.ID,
		ReceptionID: lastReception.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create product: %w", op, err)
	}

	product.ProductType = productType

	return product, nil
}

func (s *ProductUseCase) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) (*domain.Product, error) {
	const op = "products.DeleteLastProduct"

	_, err := s.pvzRepo.Get(ctx, domain.PVZ{
		ID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrPVZNotFound
		}
		return nil, fmt.Errorf("%s: failed to find pvz: %w", op, err)
	}

	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrNoReceptionIsCurrentlyInProgress
		}
		return nil, fmt.Errorf("%s: failed to find open reception: %w", op, err)
	}

	lastProduct, err := s.productRepo.GetLastProductInReception(ctx, lastReception.ID)
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrProductToDelete
		}
		return nil, fmt.Errorf("%s: failed to get last product: %w", op, err)
	}

	err = s.productRepo.DeleteProduct(ctx, lastProduct.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to delete product: %w", op, err)
	}

	return lastProduct, nil
}

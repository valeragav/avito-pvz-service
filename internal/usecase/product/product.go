package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
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

type transactionManager interface {
	RunRepeatableRead(ctx context.Context, fn func(ctx context.Context) error) error
	RunReadCommitted(ctx context.Context, fn func(ctx context.Context) error) error
}

type ProductUseCase struct {
	tm              transactionManager
	productRepo     productRepo
	receptionRepo   receptionRepo
	productTypeRepo productTypeRepo
	pvzRepo         pvzRepo
}

func New(tm transactionManager, productRepo productRepo, receptionRepo receptionRepo, productTypeRepo productTypeRepo, pvzRepo pvzRepo) *ProductUseCase {
	return &ProductUseCase{
		tm,
		productRepo,
		receptionRepo,
		productTypeRepo,
		pvzRepo,
	}
}

func (s *ProductUseCase) Create(ctx context.Context, createIn dto.ProductCreate) (*domain.Product, error) {
	const op = "products.Create"

	if err := s.checkPVZExists(ctx, createIn.PvzID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	productType, err := s.productTypeRepo.Get(ctx, domain.ProductType{Name: createIn.TypeName})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find product type '%s': %w", op, createIn.TypeName, err)
	}

	var result *domain.Product
	err = s.tm.RunRepeatableRead(ctx, func(ctx context.Context) error {
		var txErr error
		result, txErr = s.create(ctx, createIn.PvzID, productType.ID)
		return txErr
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result.ProductType = productType
	return result, nil
}

func (s *ProductUseCase) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) (*domain.Product, error) {
	const op = "products.DeleteLastProduct"

	if err := s.checkPVZExists(ctx, pvzID); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var result *domain.Product
	err := s.tm.RunReadCommitted(ctx, func(ctx context.Context) error {
		var txErr error
		result, txErr = s.deleteLastProduct(ctx, pvzID)
		return txErr
	})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *ProductUseCase) create(ctx context.Context, pvzID, typeID uuid.UUID) (*domain.Product, error) {
	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrNoReceptionIsCurrentlyInProgress
		}
		return nil, fmt.Errorf("failed to find in progress reception: %w", err)
	}

	product, err := s.productRepo.Create(ctx, domain.Product{
		DateTime:    time.Now(),
		TypeID:      typeID,
		ReceptionID: lastReception.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *ProductUseCase) deleteLastProduct(ctx context.Context, pvzID uuid.UUID) (*domain.Product, error) {
	lastReception, err := s.receptionRepo.FindByStatus(ctx, domain.ReceptionStatusInProgress, domain.Reception{
		PvzID: pvzID,
	})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrNoReceptionIsCurrentlyInProgress
		}
		return nil, fmt.Errorf("failed to find open reception: %w", err)
	}

	lastProduct, err := s.productRepo.GetLastProductInReception(ctx, lastReception.ID)
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return nil, domain.ErrProductToDelete
		}
		return nil, fmt.Errorf("failed to get last product: %w", err)
	}

	err = s.productRepo.DeleteProduct(ctx, lastProduct.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	return lastProduct, nil
}

func (s *ProductUseCase) checkPVZExists(ctx context.Context, pvzID uuid.UUID) error {
	_, err := s.pvzRepo.Get(ctx, domain.PVZ{ID: pvzID})
	if err != nil {
		if errors.Is(err, infra.ErrNotFound) {
			return domain.ErrPVZNotFound
		}
		return fmt.Errorf("failed to find pvz: %w", err)
	}
	return nil
}

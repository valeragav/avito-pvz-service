package products

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/product_types"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/products"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/pvz"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/receptions"
	"github.com/valeragav/avito-pvz-service/internal/infrastructure/storage/statuses"
)

type ProductsService struct {
	productsRepo     *products.Repository
	receptionsRepo   *receptions.Repository
	productTypesRepo *product_types.Repository
	pvzRepo          *pvz.Repository
}

func New(productsRepo *products.Repository, receptionsRepo *receptions.Repository, productTypesRepo *product_types.Repository, pvzRepo *pvz.Repository) *ProductsService {
	return &ProductsService{
		productsRepo,
		receptionsRepo,
		productTypesRepo,
		pvzRepo,
	}
}

func (s *ProductsService) Create(ctx context.Context, createIn CreateIn) (*CreateOut, error) {
	const op = "products.Create"

	lastReception, err := s.receptionsRepo.GetLastWithStatus(ctx, receptions.Receptions{
		PvzID: createIn.PvzID,
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%s: failed to find in progress reception: %w", op, err)
	}

	if lastReception.StatusName == statuses.StatusClose {
		return nil, ErrNotFoundReceptionsRepoInProgress
	}

	productType, err := s.productTypesRepo.Get(ctx, product_types.ProductTypes{Name: createIn.TypeName})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to find product type: %w", op, err)
	}

	product, err := s.productsRepo.Create(ctx, products.Products{
		DateTime:    time.Now(),
		TypeIs:      productType.ID,
		ReceptionID: lastReception.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create product: %w", op, err)
	}

	return &CreateOut{
		ID:          product.ID,
		TypeName:    productType.Name,
		ReceptionID: product.ReceptionID,
		DateTime:    product.DateTime,
	}, nil
}

func (s *ProductsService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const op = "products.DeleteLastProduct"

	_, err := s.pvzRepo.Get(ctx, pvz.Pvz{
		ID: pvzID,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to find pvz: %w", op, err)
	}

	lastReception, err := s.receptionsRepo.GetLastWithStatus(ctx, receptions.Receptions{
		PvzID: pvzID,
	})
	if err != nil {
		return fmt.Errorf("%s: failed to find open reception: %w", op, err)
	}

	if lastReception.StatusName == statuses.StatusClose {
		return ErrNotFoundReceptionsRepoInProgress
	}

	lastProduct, err := s.productsRepo.GetLastProductInReception(ctx, lastReception.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to get last product: %w", op, err)
	}

	err = s.productsRepo.DeleteProduct(ctx, lastProduct.ID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete product: %w", op, err)
	}

	return nil
}

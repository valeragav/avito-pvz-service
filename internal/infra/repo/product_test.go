package repo_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
)

type productFixture struct {
	ctx context.Context
	tx  infra.DBTX

	productRepo *repo.ProductRepository

	productType *domain.ProductType
	reception   *domain.Reception
}

func newProductFixture(t *testing.T, ctx context.Context, tx infra.DBTX) *productFixture {
	t.Helper()

	require.NoError(t, testApp.Seed(ctx, tx, SeedReceptionStatuses, SeedProductTypes))

	stableNow := time.Now().UTC().Truncate(time.Millisecond)
	productTypeRepo := repo.NewProductTypeRepository(tx)
	receptionStatusRepo := repo.NewReceptionStatusRepository(tx)
	cityRepo := repo.NewCityRepository(tx)
	pvzRepo := repo.NewPVZRepository(tx)
	receptionRepo := repo.NewReceptionRepository(tx)

	productType, err := productTypeRepo.Get(ctx, domain.ProductType{
		Name: "электроника",
	})
	require.NoError(t, err)

	receptionStatus, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{
		Name: domain.ReceptionStatusClose,
	})
	require.NoError(t, err)

	city, err := cityRepo.Create(ctx, domain.City{
		ID:   uuid.New(),
		Name: "TestCity",
	})
	require.NoError(t, err)

	pvz, err := pvzRepo.Create(ctx, domain.PVZ{
		ID:               uuid.New(),
		RegistrationDate: stableNow,
		CityID:           city.ID,
	})
	require.NoError(t, err)

	reception, err := receptionRepo.Create(ctx, domain.Reception{
		ID:       uuid.New(),
		PvzID:    pvz.ID,
		DateTime: stableNow,
		StatusID: receptionStatus.ID,
	})
	require.NoError(t, err)

	return &productFixture{
		ctx:         ctx,
		tx:          tx,
		productRepo: repo.NewProductRepository(tx),
		productType: productType,
		reception:   reception,
	}
}

func newProduct(typeID, receptionID uuid.UUID, at time.Time) domain.Product {
	return domain.Product{
		ID:          uuid.New(),
		DateTime:    at,
		TypeID:      typeID,
		ReceptionID: receptionID,
	}
}

func TestProductRepository_CreateAndGet(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		now := time.Now().UTC().Truncate(time.Millisecond)
		f := newProductFixture(t, ctx, tx)
		product := newProduct(f.productType.ID, f.reception.ID, now)

		created, err := f.productRepo.Create(ctx, product)

		require.NoError(t, err)
		require.NotNil(t, created)

		require.Equal(t, product.ID, created.ID)
		require.Equal(t, product.TypeID, created.TypeID)
		require.Equal(t, product.ReceptionID, created.ReceptionID)
		assert.WithinDuration(t, now, created.DateTime, time.Millisecond)

		got, err := f.productRepo.Get(ctx, domain.Product{ID: product.ID})

		require.NoError(t, err)
		require.Equal(t, product.ID, got.ID)
		assert.WithinDuration(t, now, got.DateTime, time.Millisecond)
	})
}

func TestProductRepository_GetLastProductInReception(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		f := newProductFixture(t, ctx, tx)
		stableNow := time.Now().UTC().Truncate(time.Millisecond)

		t1 := stableNow.Add(-2 * time.Hour)
		t2 := stableNow.Add(-1 * time.Hour)
		t3 := stableNow

		p1 := newProduct(f.productType.ID, f.reception.ID, t1)
		p2 := newProduct(f.productType.ID, f.reception.ID, t2)
		p3 := newProduct(f.productType.ID, f.reception.ID, t3)

		_, err := f.productRepo.Create(ctx, p1)
		require.NoError(t, err)

		_, err = f.productRepo.Create(ctx, p2)
		require.NoError(t, err)

		_, err = f.productRepo.Create(ctx, p3)
		require.NoError(t, err)

		last, err := f.productRepo.GetLastProductInReception(ctx, f.reception.ID)

		require.NoError(t, err)
		require.Equal(t, p3.ID, last.ID)
		assert.WithinDuration(t, t3, last.DateTime, time.Millisecond)

		t.Run("not found", func(t *testing.T) {
			got, err := f.productRepo.GetLastProductInReception(ctx, uuid.New())
			require.ErrorIs(t, err, infra.ErrNotFound)
			require.Nil(t, got)
		})
	})
}

func TestProductRepository_ListByReceptionIDsWithTypeName(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		f := newProductFixture(t, ctx, tx)

		stableNow := time.Now().UTC().Truncate(time.Millisecond)
		receptionRepo := repo.NewReceptionRepository(tx)

		secondReception, err := receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    f.reception.PvzID,
			DateTime: stableNow,
			StatusID: f.reception.StatusID,
		})
		require.NoError(t, err)

		thirdReception, err := receptionRepo.Create(ctx, domain.Reception{
			ID:       uuid.New(),
			PvzID:    f.reception.PvzID,
			DateTime: stableNow,
			StatusID: f.reception.StatusID,
		})
		require.NoError(t, err)

		p1, _ := f.productRepo.Create(ctx,
			newProduct(f.productType.ID, f.reception.ID, stableNow),
		)

		p2, _ := f.productRepo.Create(ctx,
			newProduct(f.productType.ID, secondReception.ID, stableNow),
		)

		_, _ = f.productRepo.Create(ctx,
			newProduct(f.productType.ID, thirdReception.ID, stableNow),
		)

		results, err := f.productRepo.ListByReceptionIDsWithTypeName(
			ctx,
			[]uuid.UUID{f.reception.ID, secondReception.ID},
		)

		// Assert
		require.NoError(t, err)
		require.Len(t, results, 2)

		resultMap := make(map[uuid.UUID]struct{}, len(results))
		for _, r := range results {
			resultMap[r.ID] = struct{}{}
			require.Equal(t, "электроника", r.ProductType.Name)
		}

		_, ok1 := resultMap[p1.ID]
		_, ok2 := resultMap[p2.ID]

		require.True(t, ok1)
		require.True(t, ok2)

		t.Run("empty result", func(t *testing.T) {
			empty, err := f.productRepo.ListByReceptionIDsWithTypeName(
				ctx,
				[]uuid.UUID{uuid.New()},
			)
			require.NoError(t, err)
			require.Empty(t, empty)
		})
	})
}

func TestProductRepository_DeleteProduct(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		stableNow := time.Now().UTC().Truncate(time.Millisecond)
		f := newProductFixture(t, ctx, tx)

		product, err := f.productRepo.Create(ctx,
			newProduct(f.productType.ID, f.reception.ID, stableNow),
		)
		require.NoError(t, err)

		t.Run("success", func(t *testing.T) {
			err := f.productRepo.DeleteProduct(ctx, product.ID)

			require.NoError(t, err)

			_, err = f.productRepo.Get(ctx, domain.Product{ID: product.ID})
			require.ErrorIs(t, err, infra.ErrNotFound)
		})

		t.Run("already deleted", func(t *testing.T) {
			err := f.productRepo.DeleteProduct(ctx, product.ID)
			require.ErrorIs(t, err, infra.ErrNotFound)
		})

		t.Run("non existing", func(t *testing.T) {
			err := f.productRepo.DeleteProduct(ctx, uuid.New())
			require.ErrorIs(t, err, infra.ErrNotFound)
		})
	})
}

// // =================================================================
// func TestProductRepository_CreateAndGet(t *testing.T) {
// 	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
// 		err := testApp.Seed(ctx, tx, SeedReceptionStatuses, SeedProductTypes)
// 		require.NoError(t, err)

// 		productTypRepo := repo.NewProductTypeRepository(tx)
// 		//TODO: не правильно
// 		productTyp, err := productTypRepo.Get(ctx, domain.ProductType{Name: "электроника"})
// 		require.NoError(t, err)

// 		receptionStatusRepo := repo.NewReceptionStatusRepository(tx)
// 		receptionStatus, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: domain.ReceptionStatusClose})
// 		require.NoError(t, err)

// 		cityRepo := repo.NewCityRepository(tx)
// 		city, err := cityRepo.Create(ctx, domain.City{
// 			ID:   uuid.New(),
// 			Name: "TestCity",
// 		})

// 		pvzRepo := repo.NewPvzRepository(tx)
// 		pvz, err := pvzRepo.Create(ctx, domain.PVZ{
// 			ID:               uuid.New(),
// 			RegistrationDate: time.Now(),
// 			CityID:           city.ID,
// 		})
// 		require.NoError(t, err)

// 		receptionRepo := repo.NewReceptionRepository(tx)
// 		reception, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		productRepo := repo.NewProductRepository(tx)

// 		// создаём тестовый продукт
// 		product := domain.Product{
// 			ID:          uuid.New(),
// 			DateTime:    time.Now(),
// 			TypeID:      productTyp.ID,
// 			ReceptionID: reception.ID,
// 		}

// 		created, err := productRepo.Create(ctx, product)
// 		require.NoError(t, err)
// 		require.NotNil(t, created)
// 		require.Equal(t, product.ID, created.ID)
// 		assert.WithinDuration(t, product.DateTime, created.DateTime, time.Microsecond)
// 		assert.Equal(t, product.TypeID, created.TypeID)
// 		assert.Equal(t, product.ReceptionID, created.ReceptionID)

// 		// дополнительная проверка: получаем продукт из базы напрямую
// 		createdNew, err := productRepo.Get(ctx, domain.Product{ID: created.ID})
// 		require.NoError(t, err)
// 		require.NotNil(t, createdNew)
// 		require.Equal(t, product.ID, createdNew.ID)
// 		assert.WithinDuration(t, product.DateTime, createdNew.DateTime, time.Millisecond)
// 		assert.Equal(t, product.TypeID, createdNew.TypeID)
// 		assert.Equal(t, product.ReceptionID, createdNew.ReceptionID)
// 	})
// }

// func TestProductRepository_GetLastProductInReception(t *testing.T) {
// 	WithTx(t, func(ctx context.Context, tx infra.DBTX) {

// 		err := testApp.Seed(ctx, tx, SeedReceptionStatuses, SeedProductTypes)
// 		require.NoError(t, err)

// 		productTypRepo := repo.NewProductTypeRepository(tx)
// 		//TODO: не правильно
// 		productTyp, err := productTypRepo.Get(ctx, domain.ProductType{Name: "электроника"})
// 		require.NoError(t, err)

// 		receptionStatusRepo := repo.NewReceptionStatusRepository(tx)
// 		receptionStatus, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: domain.ReceptionStatusClose})
// 		require.NoError(t, err)

// 		cityRepo := repo.NewCityRepository(tx)
// 		city, err := cityRepo.Create(ctx, domain.City{
// 			ID:   uuid.New(),
// 			Name: "TestCity",
// 		})

// 		pvzRepo := repo.NewPvzRepository(tx)
// 		pvz, err := pvzRepo.Create(ctx, domain.PVZ{
// 			ID:               uuid.New(),
// 			RegistrationDate: time.Now(),
// 			CityID:           city.ID,
// 		})
// 		require.NoError(t, err)

// 		receptionRepo := repo.NewReceptionRepository(tx)
// 		reception, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		productRepo := repo.NewProductRepository(tx)

// 		// вставляем несколько продуктов с разным DateTime
// 		products := []domain.Product{
// 			{
// 				ID:          uuid.New(),
// 				DateTime:    time.Now().Add(-2 * time.Hour),
// 				TypeID:      productTyp.ID,
// 				ReceptionID: reception.ID,
// 			},
// 			{
// 				ID:          uuid.New(),
// 				DateTime:    time.Now().Add(-1 * time.Hour),
// 				TypeID:      productTyp.ID,
// 				ReceptionID: reception.ID,
// 			},
// 			{
// 				ID:          uuid.New(),
// 				DateTime:    time.Now(),
// 				TypeID:      productTyp.ID,
// 				ReceptionID: reception.ID,
// 			},
// 		}

// 		for i := range products {
// 			created, err := productRepo.Create(ctx, products[i])
// 			require.NoError(t, err)
// 			products[i].ID = created.ID
// 		}

// 		// ---------------- GetLastProductInReception ----------------
// 		got, err := productRepo.GetLastProductInReception(ctx, reception.ID)
// 		require.NoError(t, err)
// 		require.NotNil(t, got)

// 		// должен вернуть самый последний продукт по DateTime
// 		lastProduct := products[2]

// 		require.Equal(t, lastProduct.ID, got.ID)
// 		assert.WithinDuration(t, lastProduct.DateTime, got.DateTime, time.Millisecond)
// 		assert.Equal(t, lastProduct.TypeID, got.TypeID)
// 		assert.Equal(t, lastProduct.ReceptionID, got.ReceptionID)

// 		// ---------------- Проверка NotFound ----------------
// 		got, err = productRepo.GetLastProductInReception(ctx, uuid.New())
// 		assert.ErrorIs(t, err, infra.ErrNotFound)
// 		assert.Nil(t, got)
// 	})
// }

// func TestProductRepository_ListByReceptionIDsWithTypeName(t *testing.T) {
// 	WithTx(t, func(ctx context.Context, tx infra.DBTX) {

// 		err := testApp.Seed(ctx, tx, SeedReceptionStatuses, SeedProductTypes)
// 		require.NoError(t, err)

// 		// --- получаем тип ---
// 		productTypeRepo := repo.NewProductTypeRepository(tx)
// 		productType, err := productTypeRepo.Get(ctx, domain.ProductType{Name: "электроника"})
// 		require.NoError(t, err)

// 		// --- получаем статус ---
// 		receptionStatusRepo := repo.NewReceptionStatusRepository(tx)
// 		receptionStatus, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: domain.ReceptionStatusClose})
// 		require.NoError(t, err)

// 		// --- создаём city ---
// 		cityRepo := repo.NewCityRepository(tx)
// 		city, err := cityRepo.Create(ctx, domain.City{
// 			ID:   uuid.New(),
// 			Name: "TestCity",
// 		})
// 		require.NoError(t, err)

// 		// --- создаём pvz ---
// 		pvzRepo := repo.NewPvzRepository(tx)
// 		pvz, err := pvzRepo.Create(ctx, domain.PVZ{
// 			ID:               uuid.New(),
// 			RegistrationDate: time.Now(),
// 			CityID:           city.ID,
// 		})
// 		require.NoError(t, err)

// 		// --- создаём 2 reception ---
// 		receptionRepo := repo.NewReceptionRepository(tx)

// 		reception1, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		reception2, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		productRepo := repo.NewProductRepository(tx)

// 		// --- создаём продукты для двух разных reception ---
// 		p1, err := productRepo.Create(ctx, domain.Product{
// 			ID:          uuid.New(),
// 			DateTime:    time.Now(),
// 			TypeID:      productType.ID,
// 			ReceptionID: reception1.ID,
// 		})
// 		require.NoError(t, err)

// 		p2, err := productRepo.Create(ctx, domain.Product{
// 			ID:          uuid.New(),
// 			DateTime:    time.Now(),
// 			TypeID:      productType.ID,
// 			ReceptionID: reception2.ID,
// 		})
// 		require.NoError(t, err)

// 		// --- создаём продукт для третьего reception (не должен попасть в выборку) ---
// 		reception3, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		_, err = productRepo.Create(ctx, domain.Product{
// 			ID:          uuid.New(),
// 			DateTime:    time.Now(),
// 			TypeID:      productType.ID,
// 			ReceptionID: reception3.ID,
// 		})
// 		require.NoError(t, err)

// 		// ---------------- TEST ----------------
// 		results, err := productRepo.ListByReceptionIDsWithTypeName(ctx, []uuid.UUID{reception1.ID, reception2.ID})
// 		require.NoError(t, err)
// 		require.Len(t, results, 2)

// 		// собираем ID в map для удобной проверки
// 		ids := map[uuid.UUID]struct{}{
// 			results[0].ID: {},
// 			results[1].ID: {},
// 		}

// 		_, ok1 := ids[p1.ID]
// 		_, ok2 := ids[p2.ID]

// 		require.True(t, ok1)
// 		require.True(t, ok2)

// 		// проверяем что TypeName подтянулся
// 		for _, r := range results {
// 			require.Equal(t, productType.ID, r.TypeID)
// 			require.Equal(t, "электроника", r.ProductType.Name)
// 		}

// 		// ---------------- empty result ----------------
// 		empty, err := productRepo.
// 			ListByReceptionIDsWithTypeName(ctx, []uuid.UUID{uuid.New()})

// 		require.NoError(t, err)
// 		require.Empty(t, empty)
// 	})
// }

// func TestProductRepository_DeleteProduct(t *testing.T) {
// 	WithTx(t, func(ctx context.Context, tx infra.DBTX) {

// 		err := testApp.Seed(ctx, tx, SeedReceptionStatuses, SeedProductTypes)
// 		require.NoError(t, err)

// 		// --- получаем тип ---
// 		productTypeRepo := repo.NewProductTypeRepository(tx)
// 		productType, err := productTypeRepo.Get(ctx, domain.ProductType{Name: "электроника"})
// 		require.NoError(t, err)

// 		// --- получаем статус ---
// 		receptionStatusRepo := repo.NewReceptionStatusRepository(tx)
// 		receptionStatus, err := receptionStatusRepo.Get(ctx, domain.ReceptionStatus{Name: domain.ReceptionStatusClose})
// 		require.NoError(t, err)

// 		// --- создаём city ---
// 		cityRepo := repo.NewCityRepository(tx)
// 		city, err := cityRepo.Create(ctx, domain.City{
// 			ID:   uuid.New(),
// 			Name: "TestCity",
// 		})
// 		require.NoError(t, err)

// 		// --- создаём pvz ---
// 		pvzRepo := repo.NewPvzRepository(tx)
// 		pvz, err := pvzRepo.Create(ctx, domain.PVZ{
// 			ID:               uuid.New(),
// 			RegistrationDate: time.Now(),
// 			CityID:           city.ID,
// 		})
// 		require.NoError(t, err)

// 		// --- создаём reception ---
// 		receptionRepo := repo.NewReceptionRepository(tx)
// 		reception, err := receptionRepo.Create(ctx, domain.Reception{
// 			ID:       uuid.New(),
// 			PvzID:    pvz.ID,
// 			DateTime: time.Now(),
// 			StatusID: receptionStatus.ID,
// 		})
// 		require.NoError(t, err)

// 		productRepo := repo.NewProductRepository(tx)

// 		// --- создаём продукт ---
// 		product, err := productRepo.Create(ctx, domain.Product{
// 			ID:          uuid.New(),
// 			DateTime:    time.Now(),
// 			TypeID:      productType.ID,
// 			ReceptionID: reception.ID,
// 		})
// 		require.NoError(t, err)

// 		t.Run("success delete", func(t *testing.T) {
// 			err := productRepo.DeleteProduct(ctx, product.ID)
// 			require.NoError(t, err)

// 			_, err = productRepo.Get(ctx, domain.Product{ID: product.ID})
// 			require.ErrorIs(t, err, infra.ErrNotFound)
// 		})

// 		t.Run("delete already deleted", func(t *testing.T) {
// 			err := productRepo.DeleteProduct(ctx, product.ID)
// 			require.ErrorIs(t, err, infra.ErrNotFound)
// 		})

// 		t.Run("delete non existing product", func(t *testing.T) {
// 			err := productRepo.DeleteProduct(ctx, uuid.New())
// 			require.ErrorIs(t, err, infra.ErrNotFound)
// 		})
// 	})
// }

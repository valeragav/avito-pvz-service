package repo_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/infra/repo"
)

func TestUserRepository_Create(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		userRepo := repo.NewUserRepository(tx)

		user1 := domain.User{
			ID:           uuid.New(),
			PasswordHash: "PasswordHash",
			Email:        "alice@example.com",
			Role:         domain.EmployeeRole,
		}

		created1, err := userRepo.Create(ctx, user1)
		require.NoError(t, err)
		require.NotNil(t, created1)
		assert.Equal(t, user1.ID, created1.ID)
		assert.Equal(t, user1.PasswordHash, created1.PasswordHash)
		assert.Equal(t, user1.Email, created1.Email)
		assert.Equal(t, user1.Role, created1.Role)

		user2 := domain.User{
			ID:           uuid.Nil,
			PasswordHash: "Hash",
			Email:        "bob@example.com",
			Role:         domain.EmployeeRole,
		}

		created2, err := userRepo.Create(ctx, user2)
		require.NoError(t, err)
		require.NotNil(t, created2)
		assert.NotEqual(t, uuid.Nil, created2.ID)
		assert.Equal(t, user2.PasswordHash, created2.PasswordHash)
		assert.Equal(t, user2.Email, created2.Email)
		assert.Equal(t, user2.Role, created2.Role)
	})
}

func TestUserRepository_Get(t *testing.T) {
	WithTx(t, func(ctx context.Context, tx infra.DBTX) {
		userRepo := repo.NewUserRepository(tx)

		users := []domain.User{
			{
				ID:           uuid.New(),
				PasswordHash: "Hash1",
				Email:        "alice@example.com",
				Role:         domain.EmployeeRole,
			},
			{
				ID:           uuid.New(),
				PasswordHash: "Hash2",
				Email:        "bob@example.com",
				Role:         domain.ModeratorRole,
			},
		}

		for _, u := range users {
			_, err := userRepo.Create(ctx, u)
			require.NoError(t, err)
		}

		got, err := userRepo.Get(ctx, domain.User{ID: users[0].ID})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, users[0].ID, got.ID)
		assert.Equal(t, users[0].Email, got.Email)
		assert.Equal(t, users[0].PasswordHash, got.PasswordHash)
		assert.Equal(t, users[0].Role, got.Role)

		got, err = userRepo.Get(ctx, domain.User{Email: users[1].Email})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, users[1].ID, got.ID)
		assert.Equal(t, users[1].Email, got.Email)
		assert.Equal(t, users[1].PasswordHash, got.PasswordHash)
		assert.Equal(t, users[1].Role, got.Role)

		got, err = userRepo.Get(ctx, domain.User{Role: domain.EmployeeRole})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, users[0].ID, got.ID)
		assert.Equal(t, domain.EmployeeRole, got.Role)

		_, err = userRepo.Get(ctx, domain.User{ID: uuid.New()})
		assert.ErrorIs(t, err, infra.ErrNotFound)

		_, err = userRepo.Get(ctx, domain.User{Email: "nonexistent@example.com"})
		assert.ErrorIs(t, err, infra.ErrNotFound)
	})
}

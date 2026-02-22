package seeder

import (
	"context"
	"fmt"
)

type GenericSeed[T any] struct {
	name string
	repo SeedRepository[T]
	data func() []T
}

func NewGenericSeed[T any](name string, repo SeedRepository[T], data func() []T) *GenericSeed[T] {
	return &GenericSeed[T]{name: name, repo: repo, data: data}
}

func (s *GenericSeed[T]) Name() string { return s.name }

func (s *GenericSeed[T]) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.repo.CreateBatch(ctx, s.data())
	}
}

type SeedRepository[T any] interface {
	CreateBatch(ctx context.Context, cities []T) error
}

type Seed interface {
	Name() string
	Run(ctx context.Context) error
}

type Seeder struct {
	seeds []Seed
}

func New() *Seeder {
	return &Seeder{}
}

func (s *Seeder) Add(seed Seed) {
	s.seeds = append(s.seeds, seed)
}

func (s *Seeder) Run(ctx context.Context) error {
	var combinedErr error
	for _, seed := range s.seeds {
		// Пришлось отказаться от асинхронных запросов так одновременно выполняете seed.Run,
		// а внутри они все используют один и тот же tx (infra.DBTX), что и приводит к conn busy.
		if err := seed.Run(ctx); err != nil {
			wrappedErr := fmt.Errorf("failed to run seed %q: %w", seed.Name(), err)
			if combinedErr == nil {
				combinedErr = wrappedErr
			} else {
				combinedErr = fmt.Errorf("%w; %w", combinedErr, wrappedErr)
			}
		}
	}
	return combinedErr
}

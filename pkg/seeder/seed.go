package seeder

import (
	"context"
	"fmt"
	"sync"
)

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
	var wg sync.WaitGroup
	errCh := make(chan error, len(s.seeds))

	wg.Add(len(s.seeds))

	for _, seed := range s.seeds {
		go func(seed Seed) {
			defer wg.Done()
			if err := seed.Run(ctx); err != nil {
				errCh <- fmt.Errorf("failed to run seed %q: %w", seed.Name(), err)
			}
		}(seed)
	}

	wg.Wait()
	close(errCh)

	var combinedErr error
	for err := range errCh {
		combinedErr = fmt.Errorf("%w; %w", combinedErr, err)
	}

	return combinedErr
}

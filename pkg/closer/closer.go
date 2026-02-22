package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Func func(ctx context.Context) error

type Closer struct {
	mu     sync.Mutex
	funcs  []Func
	closed bool
}

func New() *Closer {
	return &Closer{}
}

func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		panic("closer: add after close")
	}

	c.funcs = append(c.funcs, f)
}

func (c *Closer) Close(ctx context.Context) error {
	c.mu.Lock()

	// if close is run multiple times
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true

	// copy slice funcs
	funcs := append([]Func(nil), c.funcs...)
	c.mu.Unlock()

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	wg.Add(len(funcs))

	for _, f := range funcs {
		go func() {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("panic: %v", r))
					mu.Unlock()
				}
			}()

			if err := f(ctx); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return errors.Join(errs...)
}

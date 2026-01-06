package closer

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type Func func(ctx context.Context) error

type Closer struct {
	mu    sync.Mutex
	funcs []Func
}

func New() *Closer {
	return &Closer{
		funcs: make([]Func, 0),
	}
}

func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
}

func (c *Closer) Close(ctx context.Context) error {
	var (
		msgs = make([]string, 0, len(c.funcs))
		wg   sync.WaitGroup
	)

	wg.Add(len(c.funcs))

	for _, f := range c.funcs {
		go func(f Func) {
			defer wg.Done()

			if err := f(ctx); err != nil {
				c.mu.Lock()
				msgs = append(msgs, fmt.Sprintf("[!] %v", err))
				c.mu.Unlock()
			}

		}(f)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		break
	case <-ctx.Done():
		err := ctx.Err()
		if len(msgs) > 0 {
			err = fmt.Errorf(
				"shutdown finished with error(s): \n%s\nshutdown cancelled: %s",
				strings.Join(msgs, "\n"),
				ctx.Err().Error(),
			)
		} else {
			err = fmt.Errorf("shutdown cancelled: %v", ctx.Err())
		}
		return err
	}

	if len(msgs) > 0 {
		return fmt.Errorf(
			"shutdown finished with error(s): \n%s",
			strings.Join(msgs, "\n"),
		)
	}

	return nil
}

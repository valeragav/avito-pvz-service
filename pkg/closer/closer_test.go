package closer

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCloser(t *testing.T) {
	ctx := context.Background()

	t.Run("all funcs succeed", func(t *testing.T) {
		c := New()

		var counter int32
		c.Add(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
		c.Add(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})

		err := c.Close(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(2), counter)
	})

	t.Run("funcs return errors", func(t *testing.T) {
		c := New()
		c.Add(func(ctx context.Context) error { return errors.New("err1") })
		c.Add(func(ctx context.Context) error { return errors.New("err2") })

		err := c.Close(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "err1")
		require.Contains(t, err.Error(), "err2")
	})

	t.Run("funcs panic", func(t *testing.T) {
		c := New()
		c.Add(func(ctx context.Context) error { panic("panic1") })
		c.Add(func(ctx context.Context) error { panic("panic2") })

		err := c.Close(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "panic: panic1")
		require.Contains(t, err.Error(), "panic: panic2")
	})

	t.Run("mixed errors and panics", func(t *testing.T) {
		c := New()
		c.Add(func(ctx context.Context) error { panic("panic1") })
		c.Add(func(ctx context.Context) error { return errors.New("err2") })

		err := c.Close(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "panic: panic1")
		require.Contains(t, err.Error(), "err2")
	})

	t.Run("double close does nothing", func(t *testing.T) {
		c := New()
		var counter int32
		c.Add(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})

		err := c.Close(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(1), counter)

		err = c.Close(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(1), counter)
	})

	t.Run("add after close panics", func(t *testing.T) {
		c := New()
		err := c.Close(ctx)
		require.NoError(t, err)

		require.PanicsWithValue(t, "closer: add after close", func() {
			c.Add(func(ctx context.Context) error { return nil })
		})
	})
}

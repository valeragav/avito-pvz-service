package http_test

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpserver "github.com/valeragav/avito-pvz-service/internal/api/http"
)

// freeAddr возвращает свободный адрес, выделенный ОС
func freeAddr(t *testing.T) string {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := lis.Addr().String()
	require.NoError(t, lis.Close())

	return addr
}

func TestNewServer(t *testing.T) {
	t.Parallel()

	httpSrv := &nethttp.Server{Addr: "127.0.0.1:0", ReadHeaderTimeout: 5 * time.Second}
	srv := httpserver.NewServer("test", httpSrv)

	require.NotNil(t, srv)
}

func TestStartServer_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)
	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           nethttp.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.StartServer(ctx)
	}()

	// ждём пока сервер поднимется
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		assert.NoError(t, conn.Close())
		return true
	}, time.Second, 10*time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		require.NoError(t, err, "StartServer должен вернуть nil при отмене контекста")
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: StartServer не завершился после отмены контекста")
	}
}

func TestStartServer_AcceptsConnections(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/ping", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.WriteHeader(nethttp.StatusOK)
	})

	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.StartServer(ctx)
	}()

	require.Eventually(t, func() bool {
		resp, err := nethttp.Get(fmt.Sprintf("http://%s/ping", addr))
		if err != nil {
			return false
		}
		assert.NoError(t, resp.Body.Close())
		return resp.StatusCode == nethttp.StatusOK
	}, time.Second, 10*time.Millisecond)
}

func TestStartServer_ReturnsErrorOnBusyPort(t *testing.T) {
	t.Parallel()

	// занимаем порт
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, lis.Close()) })

	addr := lis.Addr().String()

	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           nethttp.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
	})

	ctx := context.Background()

	err = srv.StartServer(ctx)
	require.Error(t, err)
}

func TestShutdown_Success(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)
	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           nethttp.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.StartServer(ctx)
	}()

	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		assert.NoError(t, conn.Close())
		return true
	}, time.Second, 10*time.Millisecond)

	err := srv.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdown_WrapsError(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)

	// канал чтобы держать соединение открытым управляемо
	hangCh := make(chan struct{})

	mux := nethttp.NewServeMux()
	mux.HandleFunc("/hang", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// держим соединение пока тест не скажет закрыть
		<-hangCh
	})

	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	})

	startCtx, startCancel := context.WithCancel(context.Background())
	t.Cleanup(startCancel)

	go func() {
		_ = srv.StartServer(startCtx)
	}()

	// ждём пока сервер поднимется
	require.Eventually(t, func() bool {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		assert.NoError(t, conn.Close())
		return true
	}, time.Second, 10*time.Millisecond)

	// открываем долгое соединение — оно будет держать сервер занятым
	go func() {
		resp, err := nethttp.Get(fmt.Sprintf("http://%s/hang", addr))
		if err != nil {
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	// даём запросу дойти до хендлера
	time.Sleep(50 * time.Millisecond)

	// контекст уже отменён — Shutdown не сможет дождаться /hang и вернёт ошибку
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := srv.Shutdown(ctx)

	// разблокируем висящий хендлер
	close(hangCh)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "http server shutdown")
}

func TestShutdown_CanBeCalledWithoutStart(t *testing.T) {
	t.Parallel()

	addr := freeAddr(t)
	srv := httpserver.NewServer("test", &nethttp.Server{
		Addr:              addr,
		Handler:           nethttp.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
	})

	// сервер не запущен — Shutdown не должен паниковать
	err := srv.Shutdown(context.Background())
	assert.NoError(t, err)
}

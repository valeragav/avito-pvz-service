package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcserver "github.com/valeragav/avito-pvz-service/internal/api/grpc"
)

func TestNewServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		addr          string
		registerFuncs []grpcserver.RegisterFunc
		wantErr       bool
	}{
		{
			name:          "success: no register funcs",
			addr:          "127.0.0.1:0",
			registerFuncs: nil,
		},
		{
			name: "success: with register func",
			addr: "127.0.0.1:0",
			registerFuncs: []grpcserver.RegisterFunc{
				func(s *googlegrpc.Server) {
					// регистрируем пустой сервис
				},
			},
		},
		{
			name:    "error: invalid address",
			addr:    "invalid-addr",
			wantErr: true,
		},
		{
			name:    "error: busy port",
			addr:    "", // заполним ниже
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			addr := tt.addr

			// занимаем порт заранее чтобы получить ошибку
			if tt.name == "error: busy port" {
				lis, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				t.Cleanup(func() { assert.NoError(t, lis.Close()) })
				addr = lis.Addr().String()
			}

			srv, err := grpcserver.NewServer(
				context.Background(),
				"test-server",
				addr,
				tt.registerFuncs,
			)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, srv)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, srv)
			t.Cleanup(func() {
				_ = srv.Shutdown(context.Background())
			})
		})
	}
}

func TestNewServer_RegisterFuncsCalledOnce(t *testing.T) {
	t.Parallel()

	callCount := 0
	registerFuncs := []grpcserver.RegisterFunc{
		func(_ *googlegrpc.Server) { callCount++ },
		func(_ *googlegrpc.Server) { callCount++ },
		func(_ *googlegrpc.Server) { callCount++ },
	}

	srv, err := grpcserver.NewServer(
		context.Background(),
		"test",
		"127.0.0.1:0",
		registerFuncs,
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	assert.Equal(t, 3, callCount)
}

func TestStartServer_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	srv, err := grpcserver.NewServer(
		context.Background(),
		"test",
		"127.0.0.1:0",
		nil,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.StartServer(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
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

	srv, err := grpcserver.NewServer(
		context.Background(),
		"test",
		"127.0.0.1:0",
		nil,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.StartServer(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// пробуем подключиться к серверу
	addr := srv.Addr()
	conn, err := googlegrpc.NewClient(
		addr,
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, conn.Close()) })
}

func TestShutdown_GracefulStop(t *testing.T) {
	t.Parallel()

	srv, err := grpcserver.NewServer(
		context.Background(),
		"test",
		"127.0.0.1:0",
		nil,
	)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() {
		_ = srv.StartServer(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	err = srv.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdown_CanBeCalledMultipleTimes(t *testing.T) {
	t.Parallel()

	srv, err := grpcserver.NewServer(
		context.Background(),
		"test",
		"127.0.0.1:0",
		nil,
	)
	require.NoError(t, err)

	assert.NoError(t, srv.Shutdown(context.Background()))
	assert.NoError(t, srv.Shutdown(context.Background())) // повторный вызов не должен паниковать
}

package grpc

import (
	"github.com/valeragav/avito-pvz-service/internal/api/grpc/gen"
	"github.com/valeragav/avito-pvz-service/internal/app"
	grpc "google.golang.org/grpc"
)

func CollectRegisters(app *app.App) []RegisterFunc {
	registers := []RegisterFunc{
		func(s *grpc.Server) {
			gen.RegisterPVZServiceServer(s, NewPVZServer(app.PVZUseCase))
		},
	}

	return registers
}

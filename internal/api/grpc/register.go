package grpc

import (
	pvz_v1 "github.com/valeragav/avito-pvz-service/internal/api/grpc/gen/v1"
	"github.com/valeragav/avito-pvz-service/internal/app"
	grpc "google.golang.org/grpc"
)

func CollectRegisters(appService *app.App) []RegisterFunc {
	registers := []RegisterFunc{
		func(s *grpc.Server) {
			pvz_v1.RegisterPVZServiceServer(s, NewPVZServer(appService.PVZUseCase))
		},
	}

	return registers
}

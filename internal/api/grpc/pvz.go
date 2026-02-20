package grpc

import (
	context "context"
	"time"

	"github.com/valeragav/avito-pvz-service/internal/api/grpc/gen"
	"github.com/valeragav/avito-pvz-service/internal/usecase/pvz"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServer struct {
	gen.UnimplementedPVZServiceServer
	PVZUseCase *pvz.PVZUseCase
}

func NewPVZServer(PVZUseCase *pvz.PVZUseCase) *PVZServer {
	return &PVZServer{
		PVZUseCase: PVZUseCase,
	}
}

func (s *PVZServer) GetPVZList(
	ctx context.Context,
	req *gen.GetPVZListRequest,
) (*gen.GetPVZListResponse, error) {
	pvzs := []*gen.PVZ{
		{
			Id:               "1",
			City:             "Moscow",
			RegistrationDate: timestamppb.New(time.Now()),
		},
		{
			Id:               "2",
			City:             "Saint Petersburg",
			RegistrationDate: timestamppb.New(time.Now().Add(-24 * time.Hour)),
		},
	}

	return &gen.GetPVZListResponse{
		Pvzs: pvzs,
	}, nil
}

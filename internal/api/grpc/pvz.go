package grpc

import (
	context "context"

	pvz_v1 "github.com/valeragav/avito-pvz-service/internal/api/grpc/gen/v1"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/usecase/pvz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzUseCase *pvz.PVZUseCase
}

func NewPVZServer(PVZUseCase *pvz.PVZUseCase) *PVZServer {
	return &PVZServer{
		pvzUseCase: PVZUseCase,
	}
}

func (s *PVZServer) GetPVZList(ctx context.Context, _ *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	pvzs, err := s.pvzUseCase.ListOverview(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pvzList := pvzListToResponse(pvzs)

	pvz := &pvz_v1.GetPVZListResponse{Pvzs: pvzList}

	return pvz, nil
}

func pvzListToResponse(pvzs []*domain.PVZ) []*pvz_v1.PVZ {
	pvzList := make([]*pvz_v1.PVZ, 0, len(pvzs))
	for _, pvz := range pvzs {
		pvzList = append(pvzList, &pvz_v1.PVZ{
			Id:               pvz.ID.String(),
			RegistrationDate: timestamppb.New(pvz.RegistrationDate),
			City:             pvz.City.Name,
		})
	}

	return pvzList
}

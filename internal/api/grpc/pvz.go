package grpc

import (
	context "context"

	pvz_v1 "github.com/valeragav/avito-pvz-service/internal/api/grpc/gen/v1"
	"github.com/valeragav/avito-pvz-service/internal/domain"

	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type pvzLister interface {
	ListOverview(ctx context.Context, pvzListParams *dto.PVZListParams) ([]*domain.PVZ, error)
}

type PVZServer struct {
	pvz_v1.UnimplementedPVZServiceServer
	pvzUseCase pvzLister
}

func NewPVZServer(pVZUseCase pvzLister) *PVZServer {
	return &PVZServer{
		pvzUseCase: pVZUseCase,
	}
}

func (s *PVZServer) GetPVZList(ctx context.Context, _ *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	pvzs, err := s.pvzUseCase.ListOverview(ctx, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pvzList := pvzListToResponse(pvzs)

	return &pvz_v1.GetPVZListResponse{Pvzs: pvzList}, nil
}

func pvzListToResponse(pvzs []*domain.PVZ) []*pvz_v1.PVZ {
	pvzList := make([]*pvz_v1.PVZ, 0, len(pvzs))

	for _, pvz := range pvzs {
		var city string
		if pvz.City != nil {
			city = pvz.City.Name
		}
		pvzList = append(pvzList, &pvz_v1.PVZ{
			Id:               pvz.ID.String(),
			RegistrationDate: timestamppb.New(pvz.RegistrationDate),
			City:             city,
		})
	}

	return pvzList
}

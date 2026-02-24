package grpc

import (
	context "context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pvz_v1 "github.com/valeragav/avito-pvz-service/internal/api/grpc/gen/v1"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockPVZLister struct {
	pvzs []*domain.PVZ
	err  error
}

func (m *mockPVZLister) ListOverview(ctx context.Context, pvzListParams *dto.PVZListParams) ([]*domain.PVZ, error) {
	return m.pvzs, m.err
}

func TestGetPVZList(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)

	pvz1 := &domain.PVZ{
		ID:               uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		RegistrationDate: now,
		City:             &domain.City{Name: "Москва"},
	}
	pvz2 := &domain.PVZ{
		ID:               uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		RegistrationDate: now.Add(-24 * time.Hour),
		City:             &domain.City{Name: "Санкт-Петербург"},
	}

	tests := []struct {
		name      string
		mock      *mockPVZLister
		wantErr   bool
		wantCode  codes.Code
		wantLen   int
		wantFirst *pvz_v1.PVZ
	}{
		{
			name:    "success: returns list of pvzs",
			mock:    &mockPVZLister{pvzs: []*domain.PVZ{pvz1, pvz2}},
			wantLen: 2,
			wantFirst: &pvz_v1.PVZ{
				Id:   "11111111-1111-1111-1111-111111111111",
				City: "Москва",
			},
		},
		{
			name:    "success: empty list",
			mock:    &mockPVZLister{pvzs: []*domain.PVZ{}},
			wantLen: 0,
		},
		{
			name:    "success: nil list from usecase",
			mock:    &mockPVZLister{pvzs: nil},
			wantLen: 0,
		},
		{
			name:     "error: usecase returns error",
			mock:     &mockPVZLister{err: errors.New("db is down")},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := NewPVZServer(tt.mock)
			resp, err := srv.GetPVZList(context.Background(), &pvz_v1.GetPVZListRequest{})

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)

				st, ok := status.FromError(err)
				require.True(t, ok, "error must be a gRPC status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.GetPvzs(), tt.wantLen)

			if tt.wantFirst != nil {
				first := resp.GetPvzs()[0]
				assert.Equal(t, tt.wantFirst.GetId(), first.GetId())
				assert.Equal(t, tt.wantFirst.GetCity(), first.GetCity())
			}
		})
	}
}

func TestGetPVZList_FieldMapping(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	id := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mock := &mockPVZLister{
		pvzs: []*domain.PVZ{
			{
				ID:               id,
				RegistrationDate: fixedTime,
				City:             &domain.City{Name: "Казань"},
			},
		},
	}

	srv := NewPVZServer(mock)
	resp, err := srv.GetPVZList(context.Background(), &pvz_v1.GetPVZListRequest{})

	require.NoError(t, err)
	require.Len(t, resp.GetPvzs(), 1)

	got := resp.GetPvzs()[0]

	assert.Equal(t, id.String(), got.GetId())
	assert.Equal(t, "Казань", got.GetCity())
	// Проверяем что время корректно сконвертировано через timestamppb
	assert.Equal(t, fixedTime.Unix(), got.GetRegistrationDate().AsTime().Unix())
}

func TestGetPVZList_PreservesOrder(t *testing.T) {
	t.Parallel()

	cities := []string{"Москва", "Казань", "Новосибирск"}
	pvzs := make([]*domain.PVZ, len(cities))
	for i, city := range cities {
		pvzs[i] = &domain.PVZ{
			ID:               uuid.New(),
			RegistrationDate: time.Now(),
			City:             &domain.City{Name: city},
		}
	}

	srv := NewPVZServer(&mockPVZLister{pvzs: pvzs})
	resp, err := srv.GetPVZList(context.Background(), &pvz_v1.GetPVZListRequest{})

	require.NoError(t, err)
	require.Len(t, resp.GetPvzs(), len(cities))

	for i, city := range cities {
		assert.Equal(t, city, resp.GetPvzs()[i].GetCity())
	}
}

func TestGetPVZList_ErrorMessagePropagated(t *testing.T) {
	t.Parallel()

	const errMsg = "connection refused"
	mock := &mockPVZLister{err: errors.New(errMsg)}

	srv := NewPVZServer(mock)
	_, err := srv.GetPVZList(context.Background(), &pvz_v1.GetPVZListRequest{})

	require.Error(t, err)
	st, _ := status.FromError(err)
	assert.Contains(t, st.Message(), errMsg)
}

func TestGetPVZList_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // сразу отменяем

	mock := &mockPVZLister{err: context.Canceled}

	srv := NewPVZServer(mock)
	_, err := srv.GetPVZList(ctx, &pvz_v1.GetPVZListRequest{})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestPvzListToResponse(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	tests := []struct {
		name    string
		input   []*domain.PVZ
		wantLen int
		check   func(t *testing.T, got []*pvz_v1.PVZ)
	}{
		{
			name:    "nil input returns empty slice",
			input:   nil,
			wantLen: 0,
			check: func(t *testing.T, got []*pvz_v1.PVZ) {
				assert.NotNil(t, got)
			},
		},
		{
			name:    "empty input returns empty slice",
			input:   []*domain.PVZ{},
			wantLen: 0,
			check: func(t *testing.T, got []*pvz_v1.PVZ) {
				assert.NotNil(t, got)
			},
		},
		{
			name: "single pvz: all fields mapped correctly",
			input: []*domain.PVZ{
				{
					ID:               id1,
					RegistrationDate: fixedTime,
					City:             &domain.City{Name: "Москва"},
				},
			},
			wantLen: 1,
			check: func(t *testing.T, got []*pvz_v1.PVZ) {
				item := got[0]
				assert.Equal(t, id1.String(), item.GetId())
				assert.Equal(t, "Москва", item.GetCity())
				require.NotNil(t, item.GetRegistrationDate())
				assert.Equal(t, fixedTime.Unix(), item.GetRegistrationDate().AsTime().Unix())
			},
		},
		{
			name: "multiple pvzs: order preserved",
			input: []*domain.PVZ{
				{
					ID:               id1,
					RegistrationDate: fixedTime,
					City:             &domain.City{Name: "Москва"},
				},
				{
					ID:               id2,
					RegistrationDate: fixedTime.Add(24 * time.Hour),
					City:             &domain.City{Name: "Казань"},
				},
			},
			wantLen: 2,
			check: func(t *testing.T, got []*pvz_v1.PVZ) {
				assert.Equal(t, id1.String(), got[0].GetId())
				assert.Equal(t, "Москва", got[0].GetCity())

				assert.Equal(t, id2.String(), got[1].GetId())
				assert.Equal(t, "Казань", got[1].GetCity())
			},
		},
		{
			name: "zero time is mapped without error",
			input: []*domain.PVZ{
				{
					ID:               id1,
					RegistrationDate: time.Time{},
					City:             &domain.City{Name: "Сочи"},
				},
			},
			wantLen: 1,
			check: func(t *testing.T, got []*pvz_v1.PVZ) {
				require.NotNil(t, got[0].GetRegistrationDate())
				require.Equal(t, time.Time{}.Unix(), got[0].GetRegistrationDate().AsTime().Unix())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := pvzListToResponse(tt.input)

			require.Len(t, got, tt.wantLen)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

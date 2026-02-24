package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valeragav/avito-pvz-service/internal/domain"
	"github.com/valeragav/avito-pvz-service/internal/infra"
	"github.com/valeragav/avito-pvz-service/internal/usecase/auth/mocks"
	"github.com/valeragav/avito-pvz-service/internal/usecase/dto"
	"github.com/valeragav/avito-pvz-service/pkg/testutils"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

type authMocks struct {
	MockJwtService *mocks.MockjwtService
	MockUserRepo   *mocks.MockuserRepository
}

func newAuthMocks(t *testing.T) *authMocks {
	ctrl := gomock.NewController(t)

	return &authMocks{
		MockJwtService: mocks.NewMockjwtService(ctrl),
		MockUserRepo:   mocks.NewMockuserRepository(ctrl),
	}
}

func TestAuthUseCase_Register(t *testing.T) {
	t.Parallel()
	testutils.InitTestLogger()

	ctx := context.Background()

	password := "secret123"

	registerReq := dto.RegisterIn{
		Email:    "newuser@email.ru",
		Password: password,
		Role:     "moderator",
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	_ = hashedPassword

	type fields struct {
		name    string
		req     dto.RegisterIn
		mockFn  func(fields fields, m *authMocks)
		wantErr error
	}

	testcases := []fields{
		{
			name: "ok",
			req:  registerReq,
			mockFn: func(f fields, m *authMocks) {
				m.MockUserRepo.EXPECT().Get(ctx, domain.User{Email: f.req.Email}).
					Return(nil, infra.ErrNotFound).
					Times(1)

				m.MockUserRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, u domain.User) (*domain.User, error) {
						u.ID = uuid.New()
						return &u, nil
					}).Times(1)
			},
			wantErr: nil,
		},

		{
			name: "failed to get user",
			req:  registerReq,
			mockFn: func(fields fields, m *authMocks) {
				m.MockUserRepo.EXPECT().Get(ctx, domain.User{Email: fields.req.Email}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("auth.Register: failed to check if user exists: db error"),
		},
		{
			name: "user already exists",
			req:  registerReq,
			mockFn: func(f fields, m *authMocks) {
				m.MockUserRepo.EXPECT().
					Get(ctx, domain.User{Email: f.req.Email}).
					Return(&domain.User{Email: f.req.Email}, nil).
					Times(1)
			},
			wantErr: domain.ErrAlreadyExists,
		},

		{
			name: "repo error on Create",
			req:  registerReq,
			mockFn: func(f fields, m *authMocks) {
				m.MockUserRepo.EXPECT().
					Get(ctx, domain.User{Email: f.req.Email}).
					Return(nil, infra.ErrNotFound).
					Times(1)

				m.MockUserRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("auth.Register: to create user: db error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authMocks := newAuthMocks(t)
			tt.mockFn(tt, authMocks)

			authUseCase := New(authMocks.MockJwtService, authMocks.MockUserRepo)

			user, err := authUseCase.Register(ctx, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			if user.ID == uuid.Nil {
				require.Error(t, errors.New("id user empty"))
			}

			require.Equal(t, tt.req.Email, user.Email)
			require.Equal(t, tt.req.Role, string(user.Role))
			require.NotEmpty(t, user.PasswordHash)
		})
	}
}

func TestAuthUseCase_GenerateToken(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()

	type fields struct {
		name    string
		role    domain.Role
		token   domain.Token
		wantErr error
		mockFn  func(fields fields, m *authMocks)
	}

	testcases := []fields{
		{
			name:  "ok",
			role:  domain.ModeratorRole,
			token: domain.Token(uuid.New().String()),
			mockFn: func(f fields, m *authMocks) {
				m.MockJwtService.EXPECT().
					SignJwt(domain.UserClaims{Role: domain.ModeratorRole}).
					Return(string(f.token), nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name:  "jwt service error",
			role:  domain.ModeratorRole,
			token: "",
			mockFn: func(f fields, m *authMocks) {
				m.MockJwtService.EXPECT().
					SignJwt(domain.UserClaims{Role: domain.ModeratorRole}).
					Return("", errors.New("jwt error")).
					Times(1)
			},
			wantErr: errors.New("auth.GenerateToken: failed to generate token: jwt error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authMocks := newAuthMocks(t)
			tt.mockFn(tt, authMocks)

			authUseCase := New(authMocks.MockJwtService, authMocks.MockUserRepo)
			token, err := authUseCase.GenerateToken(tt.role)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Empty(t, token)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.token, *token)
		})
	}
}

func TestAuthUseCase_Login(t *testing.T) {
	t.Parallel()

	testutils.InitTestLogger()

	ctx := context.Background()

	password := "secret123"

	loginInReq := dto.LoginIn{
		Email:    "test@email.ru",
		Password: password,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	type fields struct {
		name           string
		hashedPassword string
		req            dto.LoginIn
		token          domain.Token
		wantErr        error
		mockFn         func(fields fields, m *authMocks)
	}

	testcases := []fields{
		{
			name:  "ok",
			req:   loginInReq,
			token: domain.Token(uuid.New().String()),
			mockFn: func(fields fields, m *authMocks) {
				m.MockUserRepo.EXPECT().
					Get(ctx, domain.User{Email: fields.req.Email}).
					Return(&domain.User{
						Email:        fields.req.Email,
						PasswordHash: string(hashedPassword),
						Role:         domain.ModeratorRole,
					}, nil).
					Times(1)

				m.MockJwtService.EXPECT().
					SignJwt(domain.UserClaims{
						Role: domain.ModeratorRole,
					}).
					Return(string(fields.token), nil).
					Times(1)
			},
			wantErr: nil,
		},
		{
			name:           "failed to get user",
			req:            loginInReq,
			token:          "",
			hashedPassword: uuid.New().String(),
			mockFn: func(fields fields, m *authMocks) {
				m.MockUserRepo.EXPECT().Get(ctx, domain.User{Email: fields.req.Email}).
					Return(nil, errors.New("db error")).
					Times(1)
			},
			wantErr: errors.New("auth.Login: failed to get user: db error"),
		},
		{
			name:  "invalid password",
			req:   loginInReq,
			token: "",
			mockFn: func(fields fields, m *authMocks) {
				m.MockUserRepo.EXPECT().
					Get(ctx, domain.User{Email: fields.req.Email}).
					Return(&domain.User{
						Email:        fields.req.Email,
						PasswordHash: string([]byte("wrong-hash")),
						Role:         domain.ModeratorRole,
					}, nil).
					Times(1)
			},
			wantErr: domain.ErrInvalidEmailOrPassword,
		},
		{
			name:  "error generate token",
			req:   loginInReq,
			token: "",
			mockFn: func(fields fields, m *authMocks) {
				m.MockUserRepo.EXPECT().
					Get(ctx, domain.User{Email: fields.req.Email}).
					Return(&domain.User{
						Email:        fields.req.Email,
						PasswordHash: string(hashedPassword),
						Role:         domain.ModeratorRole,
					}, nil).
					Times(1)

				m.MockJwtService.EXPECT().
					SignJwt(domain.UserClaims{Role: domain.ModeratorRole}).
					Return("", errors.New("jwt error")).
					Times(1)
			},
			wantErr: errors.New("auth.Login: failed to generate token: jwt error"),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			authMocks := newAuthMocks(t)
			tt.mockFn(tt, authMocks)

			authUseCase := New(authMocks.MockJwtService, authMocks.MockUserRepo)

			token, err := authUseCase.Login(ctx, tt.req)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				require.Nil(t, token)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, token)
			require.Equal(t, tt.token, *token)
		})
	}
}

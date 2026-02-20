package seeder

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockSeed struct {
	name string
	err  error
}

func (m *mockSeed) Name() string {
	return m.name
}

func (m *mockSeed) Run(ctx context.Context) error {
	return m.err
}

func TestSeeder_Run(t *testing.T) {
	ctx := context.Background()

	testcases := []struct {
		name    string
		seeds   []Seed
		wantErr bool
		errMsg  string
		checkFn func(t *testing.T, err error)
	}{
		{
			name: "all seeds succeed",
			seeds: []Seed{
				&mockSeed{name: "seed1", err: nil},
				&mockSeed{name: "seed2", err: nil},
			},
			wantErr: false,
		},
		{
			name: "one seed fails",
			seeds: []Seed{
				&mockSeed{name: "seed1", err: errors.New("fail1")},
				&mockSeed{name: "seed2", err: nil},
			},
			wantErr: true,
			errMsg:  "failed to run seed \"seed1\": fail1",
		},
		{
			name: "multiple seeds fail",
			seeds: []Seed{
				&mockSeed{name: "seed1", err: errors.New("fail1")},
				&mockSeed{name: "seed2", err: errors.New("fail2")},
			},
			wantErr: true,
			errMsg:  "failed to run seed \"seed1\": fail1; failed to run seed \"seed2\": fail2",
			checkFn: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), `failed to run seed "seed1": fail1`)
				require.Contains(t, err.Error(), `failed to run seed "seed2": fail2`)
			},
		},
		{
			name:    "no seeds",
			seeds:   []Seed{},
			wantErr: false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			seeder := New()
			for _, s := range tt.seeds {
				seeder.Add(s)
			}

			err := seeder.Run(ctx)

			if tt.checkFn != nil {
				tt.checkFn(t, err)
			}

			if tt.wantErr {
				if tt.checkFn != nil {
					tt.checkFn(t, err)
				} else {
					require.Error(t, err)
					require.EqualError(t, err, tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

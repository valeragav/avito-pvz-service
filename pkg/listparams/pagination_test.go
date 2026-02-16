package listparams

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePagination(t *testing.T) {
	defaults := Pagination{Page: 1, Limit: 20}

	testcases := []struct {
		name       string
		query      url.Values
		defaults   Pagination
		want       Pagination
		expectErr  bool
		errMessage string
	}{
		{
			name:     "no query, use defaults",
			query:    url.Values{},
			defaults: defaults,
			want:     Pagination{Page: 1, Limit: 20},
		},
		{
			name: "valid limit and page",
			query: url.Values{
				"limit": []string{"50"},
				"page":  []string{"3"},
			},
			defaults: defaults,
			want:     Pagination{Page: 3, Limit: 50},
		},
		{
			name: "limit too high",
			query: url.Values{
				"limit": []string{"200"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "limit must be between 1 and 100",
		},
		{
			name: "limit zero",
			query: url.Values{
				"limit": []string{"0"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "limit must be between 1 and 100",
		},
		{
			name: "limit negative",
			query: url.Values{
				"limit": []string{"-5"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "limit must be between 1 and 100",
		},
		{
			name: "limit not a number",
			query: url.Values{
				"limit": []string{"abc"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "limit must be a number",
		},
		{
			name: "page zero",
			query: url.Values{
				"page": []string{"0"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "page must be >= 1",
		},
		{
			name: "page negative",
			query: url.Values{
				"page": []string{"-2"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "page must be >= 1",
		},
		{
			name: "page not a number",
			query: url.Values{
				"page": []string{"abc"},
			},
			defaults:   defaults,
			expectErr:  true,
			errMessage: "page must be a number",
		},
		{
			name: "partial query - only page",
			query: url.Values{
				"page": []string{"5"},
			},
			defaults: defaults,
			want:     Pagination{Page: 5, Limit: 20},
		},
		{
			name: "partial query - only limit",
			query: url.Values{
				"limit": []string{"30"},
			},
			defaults: defaults,
			want:     Pagination{Page: 1, Limit: 30},
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParsePagination(tt.query, tt.defaults)

			if tt.expectErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.errMessage)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want.Page, got.Page)
			require.Equal(t, tt.want.Limit, got.Limit)

			expectedOffset := (got.Page - 1) * got.Limit
			require.Equal(t, expectedOffset, got.Offset())
		})
	}
}

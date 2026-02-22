package listparams

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSort(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		query      url.Values
		defaults   Sort
		want       Sort
		expectErr  bool
		errMessage string
	}{
		{
			name:     "no query, use defaults",
			query:    url.Values{},
			defaults: Sort{},
			want:     Sort{Field: "id", Order: SortAsc},
		},
		{
			name:  "custom defaults",
			query: url.Values{},
			defaults: Sort{
				Field: "name",
				Order: SortDesc,
			},
			want: Sort{Field: "name", Order: SortDesc},
		},
		{
			name: "override field",
			query: url.Values{
				"field": []string{"created_at"},
			},
			defaults: Sort{},
			want:     Sort{Field: "created_at", Order: SortAsc},
		},
		{
			name: "override order to desc",
			query: url.Values{
				"order": []string{"desc"},
			},
			defaults: Sort{},
			want:     Sort{Field: "id", Order: SortDesc},
		},
		{
			name: "override order to asc",
			query: url.Values{
				"order": []string{"asc"},
			},
			defaults: Sort{Field: "name", Order: SortDesc},
			want:     Sort{Field: "name", Order: SortAsc},
		},
		{
			name: "invalid order",
			query: url.Values{
				"order": []string{"invalid"},
			},
			defaults:   Sort{},
			expectErr:  true,
			errMessage: "invalid sort order",
		},
		{
			name: "field and order together",
			query: url.Values{
				"field": []string{"updated_at"},
				"order": []string{"desc"},
			},
			defaults: Sort{},
			want:     Sort{Field: "updated_at", Order: SortDesc},
		},
		{
			name: "order case insensitive",
			query: url.Values{
				"order": []string{"DeSc"},
			},
			defaults: Sort{},
			want:     Sort{Field: "id", Order: SortDesc},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseSort(tt.query, tt.defaults)

			if tt.expectErr {
				require.Error(t, err)
				require.EqualError(t, err, tt.errMessage)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want.Field, got.Field)
			require.Equal(t, tt.want.Order, got.Order)
		})
	}
}

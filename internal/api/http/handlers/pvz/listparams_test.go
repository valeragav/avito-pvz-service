package pvz

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parsePvzFilter(t *testing.T) {
	validStart := "2026-02-11T10:30:00Z"
	validEnd := "2026-02-12T10:30:00Z"

	startTime, _ := time.Parse(time.RFC3339, validStart)
	endTime, _ := time.Parse(time.RFC3339, validEnd)

	tests := []struct {
		name        string
		query       url.Values
		expect      PvzFilter
		expectError string
	}{
		{
			name: "valid start and end date",
			query: url.Values{
				"startDate": []string{validStart},
				"endDate":   []string{validEnd},
			},
			expect: PvzFilter{
				StartDate: &startTime,
				EndDate:   &endTime,
			},
		},
		{
			name: "only start date",
			query: url.Values{
				"startDate": []string{validStart},
			},
			expect: PvzFilter{
				StartDate: &startTime,
				EndDate:   nil,
			},
		},
		{
			name: "invalid start date",
			query: url.Values{
				"startDate": []string{"invalid"},
			},
			expectError: "invalid startDate",
		},
		{
			name: "invalid end date",
			query: url.Values{
				"endDate": []string{"invalid"},
			},
			expectError: "invalid endDate",
		},
		{
			name:  "empty query",
			query: url.Values{},
			expect: PvzFilter{
				StartDate: nil,
				EndDate:   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := parsePvzFilter(tt.query)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectError, err.Error())
				return
			}

			require.NoError(t, err)

			// Проверяем даты аккуратно (через reflect, потому что указатели)
			assert.True(t, reflect.DeepEqual(tt.expect, filter))
		})
	}
}

func Test_getParsePvzParam(t *testing.T) {
	validStart := "2026-02-11T10:30:00Z"
	validEnd := "2026-02-12T10:30:00Z"

	req := httptest.NewRequest(http.MethodGet, "/pvz?startDate="+validStart+"&endDate="+validEnd+"&limit=10&page=5", http.NoBody)

	params, err := getParsePvzParam(req)
	require.NoError(t, err)

	expectedStart, _ := time.Parse(time.RFC3339, validStart)
	expectedEnd, _ := time.Parse(time.RFC3339, validEnd)

	require.NotNil(t, params.Filter.StartDate)
	require.NotNil(t, params.Filter.EndDate)

	assert.Equal(t, expectedStart, *params.Filter.StartDate)
	assert.Equal(t, expectedEnd, *params.Filter.EndDate)

	// Проверяем pagination (зависит от реализации listparams.ParsePagination)
	assert.Equal(t, uint(10), params.Pagination.Limit)
	assert.Equal(t, uint(5), params.Pagination.Page)
}

func Test_getParsePvzParam_InvalidDate(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/pvz?startDate=invalid", http.NoBody)

	_, err := getParsePvzParam(req)
	require.Error(t, err)
	assert.Equal(t, "invalid startDate", err.Error())
}

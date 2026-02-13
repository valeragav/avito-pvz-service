package pvz

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/VaLeraGav/avito-pvz-service/pkg/listparams"
)

type PvzListParams struct {
	Filter     PvzFilter
	Pagination listparams.Pagination
}

type PvzFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

func getParsePvzParam(r *http.Request) (PvzListParams, error) {
	q := r.URL.Query()

	filter, err := parsePvzFilter(q)
	if err != nil {
		return PvzListParams{}, err
	}

	pagination, err := listparams.ParsePagination(q, listparams.Pagination{})
	if err != nil {
		return PvzListParams{}, err
	}

	return PvzListParams{
		Filter:     filter,
		Pagination: pagination,
	}, nil
}

func parsePvzFilter(q url.Values) (PvzFilter, error) {
	var f PvzFilter

	if v := q.Get("startDate"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, errors.New("invalid startDate")
		}
		f.StartDate = &t
	}

	if v := q.Get("endDate"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return f, errors.New("invalid endDate")
		}
		f.EndDate = &t
	}

	return f, nil
}

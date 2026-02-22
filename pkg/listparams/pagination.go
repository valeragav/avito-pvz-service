package listparams

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Pagination struct {
	Page  uint
	Limit uint
}

func (p Pagination) Offset() uint {
	return (p.Page - 1) * p.Limit
}

func ParsePagination(q url.Values, defaults Pagination) (Pagination, error) {
	p := defaults

	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
	if p.Page == 0 {
		p.Page = 1
	}

	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return p, errors.New("limit must be a number")
		}
		if limit < 1 || limit > MaxLimit {
			return p, fmt.Errorf("limit must be between 1 and %d", MaxLimit)
		}
		p.Limit = uint(limit)
	}

	if v := q.Get("page"); v != "" {
		page, err := strconv.Atoi(v)
		if err != nil {
			return p, errors.New("page must be a number")
		}
		if page < 1 {
			return p, errors.New("page must be >= 1")
		}
		p.Page = uint(page)
	}

	return p, nil
}

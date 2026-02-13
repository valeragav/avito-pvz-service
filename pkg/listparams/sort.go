package listparams

import (
	"errors"
	"net/url"
	"strings"
)

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type Sort struct {
	Field string
	Order SortOrder
}

func ParseSort(q url.Values, defaults Sort) (Sort, error) {
	s := defaults

	if s.Field == "" {
		s.Field = "id"
	}
	if s.Order == "" {
		s.Order = SortAsc
	}

	if field := q.Get("field"); field != "" {
		s.Field = field
	}

	if v := q.Get("order"); v != "" {
		switch strings.ToLower(v) {
		case "asc":
			s.Order = SortAsc
		case "desc":
			s.Order = SortDesc
		default:
			return s, errors.New("invalid sort order")
		}
	}

	return s, nil
}

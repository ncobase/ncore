package paging

import (
	"encoding/base64"
	"fmt"
	"time"
)

// Params holds the unified pagination parameters
type Params struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

// Result holds the pagination result
type Result[T any] struct {
	Items       []T    `json:"items"`
	Total       int    `json:"total,omitempty"`
	NextCursor  string `json:"next,omitempty"`
	HasNextPage bool   `json:"has_next"`
}

// NormalizeParams ensures that Limit is within an acceptable range
func NormalizeParams(params Params) Params {
	if params.Limit <= 0 || params.Limit > 1024 {
		params.Limit = 256
	}
	return params
}

// EncodeCursor encodes a timestamp to a cursor string
func EncodeCursor(t time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(t.Format(time.RFC3339Nano)))
}

// DecodeCursor decodes a cursor string to a timestamp
func DecodeCursor(cursor string) (time.Time, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(b))
}

// PagingFunc is a function type that implements pagination logic
type PagingFunc[T any] func(cursor string, limit int) (items []T, total int, nextCursor string, err error)

// Paginate applies pagination using the provided PagingFunc
func Paginate[T any](params Params, paginateFunc PagingFunc[T]) (*Result[T], error) {
	params = NormalizeParams(params)
	items, total, nextCursor, err := paginateFunc(params.Cursor, params.Limit+1)
	if err != nil {
		return nil, fmt.Errorf("pagination error: %v", err)
	}

	hasNextPage := false
	if len(items) > params.Limit {
		hasNextPage = true
		items = items[:params.Limit]
	}

	if items == nil {
		items = make([]T, 0)
	}

	return &Result[T]{
		Items:       items,
		Total:       total,
		NextCursor:  nextCursor,
		HasNextPage: hasNextPage,
	}, nil
}

// NoopPagingFunc is a noop paging function
func NoopPagingFunc[T any](cursor string, limit int) ([]T, int, string, error) {
	return nil, 0, "", nil
}

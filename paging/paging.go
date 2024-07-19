package paging

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type CursorProvider interface {
	GetCursorValue() string
}

type Params struct {
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
	Direction string `json:"direction"` // "forward" or "backward"
}

type Result[T CursorProvider] struct {
	Items       []T    `json:"items"`
	Total       int    `json:"total"`
	Cursor      string `json:"cursor,omitempty"`
	NextCursor  string `json:"next_cursor,omitempty"`
	PrevCursor  string `json:"prev_cursor,omitempty"`
	HasNextPage bool   `json:"has_next_page"`
	HasPrevPage bool   `json:"has_prev_page"`
}

func NormalizeParams(params Params) Params {
	if params.Limit <= 0 || params.Limit > 1024 {
		params.Limit = 256
	}
	return params
}

func EncodeCursor(value string) string {
	return base64.URLEncoding.EncodeToString([]byte(value))
}

func DecodeCursor(cursor string) (string, int64, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", 0, fmt.Errorf("failed to decode cursor: %v", err)
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid cursor format")
	}
	id := parts[0]
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid timestamp in cursor: %v", err)
	}
	return id, timestamp, nil
}

type PagingFunc[T CursorProvider] func(cursor string, limit int, direction string) (items []T, total int, err error)

func Paginate[T CursorProvider](params Params, paginateFunc PagingFunc[T]) (Result[T], error) {
	params = NormalizeParams(params)

	items, total, err := paginateFunc(params.Cursor, params.Limit+1, params.Direction)
	if err != nil {
		return Result[T]{}, fmt.Errorf("pagination error: %v", err)
	}

	hasNextPage := len(items) > params.Limit
	hasPrevPage := params.Cursor != ""

	if hasNextPage {
		items = items[:params.Limit]
	}

	var nextCursor, prevCursor string
	if len(items) > 0 {
		if params.Direction == "forward" || params.Direction == "" {
			if hasNextPage {
				nextCursor = EncodeCursor(items[len(items)-1].GetCursorValue())
			}
			if hasPrevPage {
				prevCursor = EncodeCursor(items[0].GetCursorValue())
			}
		} else {
			if hasNextPage {
				prevCursor = EncodeCursor(items[0].GetCursorValue())
			}
			if hasPrevPage {
				nextCursor = EncodeCursor(items[len(items)-1].GetCursorValue())
			}
			// Reverse the items for backward pagination
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	return Result[T]{
		Items:       items,
		Total:       total,
		NextCursor:  nextCursor,
		PrevCursor:  prevCursor,
		HasNextPage: hasNextPage,
		HasPrevPage: hasPrevPage,
	}, nil
}

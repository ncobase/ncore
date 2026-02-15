// Package paging provides cursor-based pagination utilities for efficient
// large dataset traversal in web APIs.
//
// Cursor-based pagination is superior to offset-based pagination for:
//   - Large datasets (no performance degradation with deep pages)
//   - Real-time data (handles insertions/deletions gracefully)
//   - Consistent results (no duplicates or missing items)
//   - Scalability (constant-time lookups)
//
// # Basic Usage
//
// Create pagination parameters from request:
//
//	params := &paging.Params{
//	    Cursor: r.URL.Query().Get("cursor"),
//	    Limit:  20,
//	}
//
// Execute paginated query and build response:
//
//	items, nextCursor := queryItems(params.Cursor, params.Limit)
//
//	result := paging.NewResult(items, nextCursor, params.Limit)
//	// Returns: {items: [...], next_cursor: "...", has_more: true}
//
// # Cursor Encoding
//
// The package handles cursor encoding/decoding automatically:
//
//	// Encode a cursor value (e.g., last item ID)
//	cursor := paging.EncodeCursor("last-item-id")
//
//	// Decode cursor for querying
//	value, err := paging.DecodeCursor(cursor)
//	if err != nil {
//	    return paging.ErrInvalidCursor
//	}
//
// # Custom Cursor Providers
//
// Implement CursorProvider interface for custom cursor logic:
//
//	type Item struct {
//	    ID        string
//	    Timestamp int64
//	}
//
//	func (i *Item) GetCursorValue() string {
//	    return fmt.Sprintf("%d:%s", i.Timestamp, i.ID)
//	}
//
// # Response Structure
//
// Standard pagination response:
//
//	{
//	  "items": [...],           // Array of results
//	  "next_cursor": "...",     // Cursor for next page
//	  "has_more": true,         // Whether more results exist
//	  "limit": 20               // Page size
//	}
//
// # Best Practices
//
//   - Use composite cursors (timestamp + ID) for ordering stability
//   - Set reasonable limit defaults (10-100 items)
//   - Always validate and sanitize cursor values
//   - Return empty next_cursor when no more results
//   - Use indexes on cursor fields for optimal performance
package paging

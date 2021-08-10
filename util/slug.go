package util

import "github.com/gosimple/slug"

// UnicodeSlug generate slug from unicode string,
func UnicodeSlug(s string) string {
	return slug.Make(s)
}

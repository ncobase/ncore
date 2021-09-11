package util

import "github.com/gosimple/slug"

// UnicodeSlug -  Make slug from unicode string,
func UnicodeSlug(s string) string {
	return slug.Make(s)
}

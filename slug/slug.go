package slug

import "github.com/gosimple/slug"

// Unicode generate slug from unicode string,
func Unicode(s string) string {
	return slug.Make(s)
}

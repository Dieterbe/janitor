package janitor

import "strings"

// Subpath returns whether p2 is contained within p1.
// see unit test for details.
func SubPath(p1, p2 string) bool {
	// force a path separator onto p1 to avoid foo/bar being considered the parent of foo/bar-baz
	if !strings.HasSuffix(p1, "/") {
		p1 += "/"
	}
	return strings.HasPrefix(p2, p1)
}

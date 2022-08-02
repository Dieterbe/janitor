package janitor

import "strings"

// Subpath returns whether p2 is contained within p1.
// paths are assumed to be relative like "foo/bar", with the root expressed as "."
// see unit test for details.
func SubPath(p1, p2 string) bool {
	// force a path separator onto p1 to avoid foo/bar being considered the parent of foo/bar-baz
	if !strings.HasSuffix(p1, "/") {
		p1 += "/"
	}

	// if p2 is not the root path, prepend the prefix, such that it would be correctly recognized as subpath of the root (".", or now updated to "./")
	if p2 != "." {
		p2 = "./" + p2
	}
	// if p1 is not the root path, prepend the prefix, in accordance to the adjustment to p2, to not break the prefix check
	if p1 != "./" {
		p1 = "./" + p1
	}
	return strings.HasPrefix(p2, p1)
}

// Subpath returns whether p2 is contained within, or equals p1.
// paths are assumed to be relative like "foo/bar", with the root expressed as "."
// see unit test for details.
func SubPathInclusive(p1, p2 string) bool {
	// force a path separator onto p1 to avoid foo/bar being considered the parent of foo/bar-baz
	if !strings.HasSuffix(p1, "/") {
		p1 += "/"
	}

	// because we did the above, we need this to not break the equality check
	// it won't break/affect the prefix check
	if !strings.HasSuffix(p2, "/") {
		p2 += "/"
	}

	// if p2 is not the root path, prepend the prefix, such that it would be correctly recognized as subpath of the root (".", or now updated to "./")
	if p2 != "./" {
		p2 = "./" + p2
	}

	// if p1 is not the root path, prepend the prefix, in accordance to the adjustment to p2, to not break the prefix check
	if p1 != "./" {
		p1 = "./" + p1
	}
	return p1 == p2 || strings.HasPrefix(p2, p1)
}

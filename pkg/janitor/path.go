package janitor

import "strings"

// Child returns whether p2 is a a child of p1 (whether path p2 is contained within p1)
// paths are assumed to be relative like "foo/bar", with the root expressed as "."
// see unit test for details.
func Child(p1, p2 string) bool {
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

// BothChildren checks whether p3 and p4 are children of p1 and p2 (or p2 and p1)
func BothChildren(p1, p2, p3, p4 string) bool {
	return false ||
		Child(p1, p3) && Child(p2, p4) ||
		Child(p1, p4) && Child(p2, p3)
}

// AChildAMatch checks whether p3/p4 is a child and an equal to p1/p2
func AChildAMatch(p1, p2, p3, p4 string) bool {
	return false ||
		Child(p1, p3) && p2 == p4 ||
		Child(p1, p4) && p2 == p3 ||
		Child(p2, p3) && p1 == p4 ||
		Child(p2, p4) && p1 == p3
}

// AChildAParent checks whether p3 and p4 have a child and a parent in p1 and p2.
func AChildAParent(p1, p2, p3, p4 string) bool {
	return false ||
		Child(p1, p3) && Child(p4, p2) ||
		Child(p1, p4) && Child(p3, p2) ||
		Child(p2, p3) && Child(p4, p1) ||
		Child(p2, p4) && Child(p3, p1)
}

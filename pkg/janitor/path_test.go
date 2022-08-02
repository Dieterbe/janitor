package janitor

import (
	"fmt"
	"testing"
)

func TestSubPath(t *testing.T) {
	for i, test := range []struct {
		p1, p2              string
		expSubPath          bool
		expSubPathInclusive bool
	}{
		{
			p1:                  ".",
			p2:                  "foo",
			expSubPath:          true,
			expSubPathInclusive: true,
		},
		{
			p1:                  "foo",
			p2:                  "foo/bar",
			expSubPath:          true,
			expSubPathInclusive: true,
		},
		{
			p1:                  "foo/bar",
			p2:                  "foo/bar.zip", // if we were to do a simple strings.HasPrefix check, it would fail here.
			expSubPath:          false,
			expSubPathInclusive: false,
		},
		{
			p1:                  "foo/bar",
			p2:                  "foo/bar",
			expSubPath:          false,
			expSubPathInclusive: true,
		},
		{
			p1:                  "foo/bar/",
			p2:                  "foo/bar",
			expSubPath:          false,
			expSubPathInclusive: true,
		},
		{
			p1:                  "dir1.zip/dir1",
			p2:                  "dir1",
			expSubPath:          false,
			expSubPathInclusive: false,
		},
		{
			p1:                  "dir1",
			p2:                  "dir1.zip/dir1",
			expSubPath:          false,
			expSubPathInclusive: false,
		},
	} {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			if got := SubPath(test.p1, test.p2); got != test.expSubPath {
				t.Errorf("SubPath(%q, %q) = %v, want %v", test.p1, test.p2, got, test.expSubPath)
			}
			if got := SubPathInclusive(test.p1, test.p2); got != test.expSubPathInclusive {
				t.Errorf("SubPathInclusive(%q, %q) = %v, want %v", test.p1, test.p2, got, test.expSubPathInclusive)
			}
		})
	}
}

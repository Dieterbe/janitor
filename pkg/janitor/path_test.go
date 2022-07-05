package janitor

import (
	"fmt"
	"testing"
)

func TestSubPath(t *testing.T) {
	for i, test := range []struct {
		p1, p2 string
		exp    bool
	}{
		{
			p1:  "/",
			p2:  "/foo",
			exp: true,
		},
		{
			p1:  "/foo",
			p2:  "/foo/bar",
			exp: true,
		},
		{
			p1:  "/foo/bar",
			p2:  "/foo/bar.zip", // if we were to do a simple strings.HasPrefix check, it would fail here.
			exp: false,
		},
		{
			p1:  "/home/dieter/go/src/github.com/Dieterbe/janitor/pkg/janitor/testdata/dir1.zip/dir1",
			p2:  "/home/dieter/go/src/github.com/Dieterbe/janitor/pkg/janitor/testdata/dir1",
			exp: false,
		},
		{
			p1:  "/home/dieter/go/src/github.com/Dieterbe/janitor/pkg/janitor/testdata/dir1",
			p2:  "/home/dieter/go/src/github.com/Dieterbe/janitor/pkg/janitor/testdata/dir1.zip/dir1",
			exp: false,
		},
	} {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			if got := SubPath(test.p1, test.p2); got != test.exp {
				t.Errorf("Distinct(%q, %q) = %v, want %v", test.p1, test.p2, got, test.exp)
			}
		})
	}
}

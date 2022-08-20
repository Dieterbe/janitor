package janitor

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSha256FingerPrint(t *testing.T) {

	exp := FilePrint{
		Path: "foo",
		Size: 3,
		Hash: FooHash,
	}

	got, err := Sha256FingerPrint("foo", strings.NewReader("foo"))
	if err != nil {
		t.Errorf("Fingerprint unexpected error %v", err)
	}

	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("Fingerprint mismatch for foo (-want +got):\n%s", diff)
	}

	exp = FilePrint{
		Path: "bar",
		Size: 3,
		Hash: BarHash,
	}
	got, err = Sha256FingerPrint("bar", strings.NewReader("bar"))
	if err != nil {
		t.Errorf("Fingerprint unexpected error %v", err)
	}

	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("Fingerprint mismatch for bar (-want +got):\n%s", diff)
	}

}

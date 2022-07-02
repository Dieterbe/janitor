package hdc

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSha256FingerPrint(t *testing.T) {

	// echo -n foo | sha256sum
	// 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae  -
	// echo -n bar | sha256sum
	// fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9  -
	fooSlice, err := hex.DecodeString("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	perr(err)
	barSlice, err := hex.DecodeString("fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	perr(err)

	exp := FilePrint{
		Path: "foo",
		Size: 3,
	}
	copy(exp.Hash[:], fooSlice)

	got := Sha256FingerPrint("foo", strings.NewReader("foo"))

	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("Fingerprint mismatch for foo (-want +got):\n%s", diff)
	}

	exp = FilePrint{
		Path: "bar",
		Size: 3,
	}
	copy(exp.Hash[:], barSlice)
	got = Sha256FingerPrint("bar", strings.NewReader("bar"))
	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("Fingerprint mismatch for bar (-want +got):\n%s", diff)
	}

}

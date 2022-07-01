package hdc

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func xor(a, b [32]byte) [32]byte {
	for i := 0; i < 32; i++ {
		a[i] ^= b[i]
	}
	return a
}

func TestFingerprinter(t *testing.T) {

	// echo -n foo | sha256sum
	// 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae  -
	// echo -n bar | sha256sum
	// fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9  -
	fooSlice, err := hex.DecodeString("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	perr(err)
	barSlice, err := hex.DecodeString("fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	perr(err)
	var foo, bar [32]byte
	copy(foo[:], fooSlice)
	copy(bar[:], barSlice)

	var fp Sha256FingerPrinter
	fp.Add("foo", strings.NewReader("foo"))

	exp := Sha256FingerPrinter{
		Prints: []FilePrint{
			{
				Path: "foo",
				Hash: foo,
			},
		},
	}

	if diff := cmp.Diff(exp, fp); diff != "" {
		t.Errorf("Fingerprint mismatch after adding foo (-want +got):\n%s", diff)
	}

	fp.Add("bar", strings.NewReader("bar"))
	exp.Prints = append(exp.Prints, FilePrint{Path: "bar", Hash: bar})

	if diff := cmp.Diff(exp, fp); diff != "" {
		t.Errorf("Fingerprint mismatch after adding bar (-want +got):\n%s", diff)
	}

}

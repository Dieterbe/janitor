package hdc

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO unit test that given a fs.FS, this structure is created. maybe a test for walk()
// note that the expected prints are ordered by their hash, which is different from the order of the DirPrint structure
func TestDirPrint(t *testing.T) {

	h1S, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	perr(err)
	h2S, err := hex.DecodeString("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	perr(err)
	h3S, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	perr(err)
	h4S, err := hex.DecodeString("fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	perr(err)
	h5S, err := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	perr(err)

	var h1, h2, h3, h4, h5 [32]byte
	copy(h1[:], h1S)
	copy(h2[:], h2S)
	copy(h3[:], h3S)
	copy(h4[:], h4S)
	copy(h5[:], h5S)

	dp := DirPrint{
		Path: "root",
		Files: []FilePrint{
			{
				Path: "a",
				Hash: h2,
			},
			{
				Path: "z",
				Hash: h1,
			},
		},
		Dirs: []DirPrint{
			{
				Path: "b",
				Files: []FilePrint{
					{
						Path: "1",
						Hash: h4,
					},
					{
						Path: "2",
						Hash: h3,
					},
					{
						Path: "3",
						Hash: h5,
					},
				},
				Dirs: nil,
			},
		},
	}

	var got []FilePrint
	dpi := dp.Iterator()
	for dpi.Next() {
		v, _ := dpi.Value()
		got = append(got, v)
	}

	exp := []FilePrint{
		{
			Path: "root/z",
			Hash: h1,
		},
		{
			Path: "root/a",
			Hash: h2,
		},
		{
			Path: "root/b/2",
			Hash: h3,
		},
		{
			Path: "root/b/1",
			Hash: h4,
		},
		{
			Path: "root/b/3",
			Hash: h5,
		},
	}

	if diff := cmp.Diff(exp, got); diff != "" {
		t.Errorf("DirPrint iteration mismatch (-want +got):\n%s", diff)
	}

}

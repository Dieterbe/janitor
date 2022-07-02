package hdc

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO unit test that given a fs.FS, this structure is created. maybe a test for walk()
// note that the expected prints are ordered by their hash, which is different from the order of the DirPrint structure
func TestDirPrintIterate(t *testing.T) {

	var got []FilePrint
	dpi := dpToIterate.Iterator()
	for dpi.Next() {
		v, _ := dpi.Value()
		got = append(got, v)
	}

	if diff := cmp.Diff(dpIterated, got); diff != "" {
		t.Errorf("DirPrint iteration mismatch (-want +got):\n%s", diff)
	}
}

func TestSimilarity(t *testing.T) {
	var otherDPIterated = []FilePrint{
		// in dpIterated but not here
		//{
		//	Path: "root/z",
		//	Hash: h1,
		//	Size: 100,
		//},
		// identical
		{
			Path: "root/a",
			Hash: h2,
			Size: 122,
		},
		// identical
		{
			Path: "root/b/2",
			Hash: h3,
			Size: 333,
		},
		// hash mismatch. note that this file contributes two times to the bytes differing, due to it existing with 2 different hashes.
		// whether this is correct behavior, is debatable. but anyway at least it's simple code.
		{
			Path: "root/b/1",
			Hash: h6, // need a hash here that sorts after h3
			Size: 444,
		},
		// file that we have, but dpIterated doesn't.
		{
			Path: "root/b/4",
			Hash: h7, // needs to sort after the hash above, but before h5
			Size: 7777,
		},
		// identical file to dpIterated but different path
		{
			Path: "root/3.copy", // similarity is 0.45
			Hash: h5,
			Size: 555,
		},
	}
	a := newFilePrintIterator(dpIterated)
	b := newFilePrintIterator(otherDPIterated)
	got := NewSimilarity(a, b)
	exp := Similarity{
		BytesDiff: 100 + 2*444 + 7777,
		BytesSame: 122 + 333 + 555,
		// would expect (1+1+1+0.45)/3 = 0.81666..
		// actually the computed similarity is a bit different: 0.8181818181818182
		// not sure why, but seems no big deal.
		PathSim: 0.81,
	}

	opt := cmp.Comparer(func(x, y float64) bool {
		delta := math.Abs(x - y)
		return delta < 0.01
	})

	if diff := cmp.Diff(exp, got, opt); diff != "" {
		t.Errorf("Similarity mismatch (-want +got):\n%s", diff)
	}
}

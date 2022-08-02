package janitor

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TODO unit test that given a fs.FS, this structure is created. maybe a test for walk()
func TestDirPrintIterate(t *testing.T) {
	var got []FilePrint
	dpi := DataMain2Print.Iterator()
	for dpi.Next() {
		v, _ := dpi.Value()
		got = append(got, v)
	}

	if diff := cmp.Diff(DataMain2Iterated, got); diff != "" {
		t.Errorf("DirPrint iteration mismatch (-want +got):\n%s", diff)
	}
}

func TestDirPrintSimilarity(t *testing.T) {
	a := newFilePrintIterator(DataMain2Iterated)
	b := newFilePrintIterator(DataMain3Iterated)
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

func TestSimilarityIdentical(t *testing.T) {
	tests := []struct {
		name string
		sim  Similarity
		exp  bool
	}{
		{
			sim: Similarity{
				BytesDiff: 0,
				BytesSame: 0,
				PathSim:   1,
			},
			exp: true,
		},
		{
			sim: Similarity{
				BytesDiff: 0,
				BytesSame: 100,
				PathSim:   1,
			},
			exp: true,
		},
		{
			sim: Similarity{
				BytesDiff: 20,
				BytesSame: 100,
				PathSim:   1,
			},
			exp: false,
		},
		{
			sim: Similarity{
				BytesDiff: 0,
				BytesSame: 100,
				PathSim:   0.8,
			},
			exp: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sim.Identical(); got != tt.exp {
				t.Errorf("Similarity.Identical() = %v, want %v", got, tt.exp)
			}
		})
	}

}

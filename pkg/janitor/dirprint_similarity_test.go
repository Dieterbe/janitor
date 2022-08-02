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
		PathSim:   float64(2) / 3,
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

func TestSimilaritySimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        Similarity
		b        Similarity
		expBytes int
		expLess  bool
	}{
		{
			name: "empty",
			a: Similarity{
				BytesDiff: 0,
				BytesSame: 0,
				PathSim:   1,
			},
			b: Similarity{
				BytesDiff: 0,
				BytesSame: 0,
				PathSim:   1,
			},
			expBytes: 0,
			expLess:  true,
		},
		{name: "same ratio of diff/same, but different pathsim",
			a: Similarity{
				BytesDiff: 1000,
				BytesSame: 1000 * 1000,
				PathSim:   0.8,
			},
			b: Similarity{
				BytesDiff: 1,
				BytesSame: 1000,
				PathSim:   0.9,
			},
			expBytes: 0,
			expLess:  true,
		},
		{name: "smaller ratio of diff/same bytes",
			a: Similarity{
				BytesDiff: 10,
				BytesSame: 1000 * 1000,
				PathSim:   1,
			},
			b: Similarity{
				BytesDiff: 10,
				BytesSame: 100 * 1000,
				PathSim:   1,
			},
			expBytes: -1,
			expLess:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.CompareBytes(tt.b); got != tt.expBytes {
				t.Errorf("case %v - CompareBytes() = %v, want %v", tt.name, got, tt.expBytes)
			}
			if got := tt.a.Less(tt.b); got != tt.expLess {
				t.Errorf("case %v - Less() = %v, want %v", tt.name, got, tt.expLess)
			}
		})
	}

}

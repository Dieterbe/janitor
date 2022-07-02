package hdc

import (
	"bytes"
	"fmt"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type Similarity struct {
	BytesSame int64
	BytesDiff int64
	PathSim   float64 // (average of all path similarities for content with a hash match)
}

func NewSimilarity(a, b Iterator) Similarity {
	var sim Similarity
	var pathsCompared int

	a.Next()
	b.Next()

	for {
		av, aok := a.Value()
		bv, bok := b.Value()

		if !aok && !bok {
			break
		}

		if aok && !bok {
			sim.BytesDiff += av.Size
			a.Next()
			continue
		}

		if !aok && bok {
			sim.BytesDiff += bv.Size
			b.Next()
			continue
		}

		// aok && bok

		if bytes.Compare(av.Hash[:], bv.Hash[:]) < 0 {
			sim.BytesDiff += av.Size
			a.Next()
			continue
		}

		if bytes.Compare(av.Hash[:], bv.Hash[:]) > 0 {
			sim.BytesDiff += bv.Size
			b.Next()
			continue
		}

		// bytes.Compare(av.Hash[:], bv.Hash[:]) == 0

		sim.BytesSame += av.Size // NOTE: we assume here that the files are the same size
		similarity := strutil.Similarity(av.Path, bv.Path, metrics.NewHamming())
		fmt.Printf("similarity between %q and %q is %.2f\n", av.Path, bv.Path, similarity)
		sim.PathSim += similarity
		a.Next()
		b.Next()

		pathsCompared++
	}

	if pathsCompared > 0 {
		sim.PathSim = sim.PathSim / float64(pathsCompared)
	}
	return sim
}

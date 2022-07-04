package hdc

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type Similarity struct {
	BytesSame int64
	BytesDiff int64
	PathSim   float64 // (average of all path similarities for content with a hash match)
}

func (s Similarity) String() string {
	// TODO could this overflow?
	return fmt.Sprintf("<Similarity bytes=%.2f path=%.2f>", float64(s.BytesSame)/float64(s.BytesSame+s.BytesDiff), s.PathSim)
}

func (s1 Similarity) Less(s2 Similarity) bool {
	// calculate which similarity is higher, where similarity is defined as same / (same+diff)
	// but we simplify the formula to remove float conversions and rounding errors.
	// if s1.BytesSame / (s1.BytesSame + s1.BytesDiff) < s2.BytesSame / (s2.BytesSame + s2.BytesDiff) {
	// if (s1.BytesSame + s1.BytesDiff) / s1.BytesSame > (s2.BytesSame + s2.BytesDiff) / s2.BytesSame { // invert both sides
	// if (s1.BytesSame + s1.BytesDiff)* s2.BytesSame > (s2.BytesSame + s2.BytesDiff) * s1.BytesSame { // multiply both sides by s1*BytesSame*s2.BytesSame
	// if s1.BytesSame*s2.BytesSame + s1.BytesDiff* s2.BytesSame > s2.BytesSame*s1.BytesSame + s2.BytesDiff* s1.BytesSame { // work out
	// if s1.BytesDiff*s2.BytesSame > s2.BytesDiff*s1.BytesSame { // remove common term
	if s1.BytesDiff*s2.BytesSame > s2.BytesDiff*s1.BytesSame {
		return true
	}
	if s1.BytesDiff*s2.BytesSame < s2.BytesDiff*s1.BytesSame {
		return false
	}
	return s1.PathSim < s2.PathSim
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

		// we assume here that the files are the same size
		// specifically, that sha256 hashes don't collide.
		sim.BytesSame += av.Size
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

type PairSim struct {
	Path1 string
	Path2 string
	Sim   Similarity
}

func GetPairSims(all map[string]DirPrint) []PairSim {
	type seenKey struct {
		p1 string
		p2 string
	}
	seen := make(map[seenKey]struct{})
	var pairSims []PairSim

	for k1, dp1 := range all {
		for k2, dp2 := range all {
			if k1 == k2 {
				continue
			}
			if strings.HasPrefix(k1, k2) || strings.HasPrefix(k2, k1) {
				continue
			}
			sk := seenKey{k1, k2}
			if k1 > k2 {
				sk.p1, sk.p2 = sk.p2, sk.p1
			}
			if _, ok := seen[sk]; ok {
				// this is a pretty naive way to avoid duplicates
				// would it be better to only iterate sub-ranges of the keys? that would require getting all keys first
				// for now, this is good enough
				continue
			}
			seen[sk] = struct{}{}

			it1 := dp1.Iterator()
			it2 := dp2.Iterator()
			p := PairSim{
				Path1: sk.p1,
				Path2: sk.p2,
				Sim:   NewSimilarity(it1, it2),
			}
			pairSims = append(pairSims, p)
		}
	}

	// sort pairSims by Similarity
	sort.Slice(pairSims, func(i, j int) bool {
		return pairSims[i].Sim.Less(pairSims[j].Sim)
	})
	return pairSims

}

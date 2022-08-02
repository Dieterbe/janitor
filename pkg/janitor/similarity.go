package janitor

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type Similarity struct {
	BytesSame int64
	BytesDiff int64
	PathSim   float64 // (average of all path similarities for content with a hash match)
}

func (s Similarity) Identical() bool {

	// this is... probably good enough?
	return s.BytesDiff == 0 && s.PathSim >= 0.99
}

func (s Similarity) ContentSimilarity() float64 {
	// TODO could this overflow?
	return float64(s.BytesSame) / float64(s.BytesSame+s.BytesDiff)
}

func (s Similarity) String() string {
	return fmt.Sprintf("<Similarity bytes=%.2f path=%.2f>", s.ContentSimilarity(), s.PathSim)
}

// CompareBytes returns -1 if s1 has lower similarity amongst it bytes than s2, +1 if it's opposite,
// or 0 if they are equivalent.  Similarity is defined as same / (same+diff)
func (s1 Similarity) CompareBytes(s2 Similarity) int {
	// First we do an algebraic simplification of the formula to remove float conversions and rounding errors.
	//                                                                                                  # starting formula
	// if s1.BytesSame / (s1.BytesSame + s1.BytesDiff) < s2.BytesSame / (s2.BytesSame + s2.BytesDiff) {
	//                                                                                                  # invert both sides
	// if (s1.BytesSame + s1.BytesDiff) / s1.BytesSame > (s2.BytesSame + s2.BytesDiff) / s2.BytesSame {
	//                                                                                                  # multiply both sides by s1.BytesSame*s2.BytesSame
	// if (s1.BytesSame + s1.BytesDiff) * s2.BytesSame > (s2.BytesSame + s2.BytesDiff) * s1.BytesSame {
	//                                                                                                  # work out (expand) the multiplications to separate terms
	// if s1.BytesSame*s2.BytesSame + s1.BytesDiff*s2.BytesSame > s2.BytesSame*s1.BytesSame + s2.BytesDiff* s1.BytesSame
	//                                                                                                  # remove common term
	// if s1.BytesDiff*s2.BytesSame > s2.BytesDiff*s1.BytesSame {

	if s2.BytesSame*s1.BytesDiff > s1.BytesSame*s2.BytesDiff {
		return -1
	}
	if s2.BytesSame*s1.BytesDiff > s1.BytesSame*s2.BytesDiff {
		return 1
	}
	return 0
}

// Less returns whether s1 is less similar than s2.
func (s1 Similarity) Less(s2 Similarity) bool {
	diff := s1.CompareBytes(s2)
	if diff < 0 {
		return true
	}
	if diff > 0 {
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
		//fmt.Printf("similarity between %q and %q is %.2f\n", av.Path, bv.Path, similarity)
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

// keys are paths within an implicit walkPath
func GetPairSims(all map[string]DirPrint) []PairSim {
	type seenKey struct {
		p1 string
		p2 string
	}
	seen := make(map[seenKey]PairSim)      // seen paths that aren't identical.
	seenIdent := make(map[seenKey]PairSim) // seen paths that are identical.
	var pairSims []PairSim

	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}

	// make sure parent directories come before children directories.
	// will allow us to skip over testing children if parents are identical (see below).
	sort.Strings(keys)

	for _, k1 := range keys {
		dp1 := all[k1]
	Loop2:
		for _, k2 := range keys {
			dp2 := all[k2]

			// don't compare to self
			if k1 == k2 {
				continue
			}

			// don't compare to a subdirectory of self.
			// e.g. pointless to compare foo/bar/baz to foo/bar , it's obvious they will have some similarity, but not due to redundancy of data warranting cleanup.
			if SubPath(k1, k2) || SubPath(k2, k1) {
				continue
			}

			// don't compare these directories if they were already compared.
			// (due to double loop, each pair will be compared twice).
			// this is pretty naive. perhaps later we can do something more clever like
			// only iterate subranges of the keys.
			sk := seenKey{p1: k1, p2: k2}
			if k1 > k2 {
				sk = seenKey{p1: k2, p2: k1}
			}
			if _, ok := seen[sk]; ok {
				continue
			}
			if _, ok := seenIdent[sk]; ok {
				continue
			}

			// If we already know that either:
			// - a pair of (grand)parents of dp1/dp2 are identical.
			// - a pair of a (grand)parent of dp1 and dp2 (or vice versa) are identical.
			// then there is no value in reporting when:
			// a) dp1 and dp2 are identical
			// b) non-symmetrical paths within dp1 and dp2 are not identical (this is debatable, but what we go with for now)
			// For example:
			// If a and b are identical,
			// a) we certainly don't care that a/foo are b/foo identical.
			// b) we likely don't care that a/somedir and b/some-other-dir are not identical.
			// Note: We probably still want to know if a/somedir and a/some-other-dir happen to also be identical,
			// so we don't filter out that case.
			// This also catches the scenario of this exact pair already having been checked,
			// which is not necessary, but is harmless and keeps the code simple.

			for ident := range seenIdent {
				a := SubPathInclusive(ident.p1, k1) && SubPathInclusive(ident.p2, k2)
				b := SubPathInclusive(ident.p1, k2) && SubPathInclusive(ident.p2, k1)

				if a || b {

					fmt.Println("DIE parent", ident, "identical. CAN SKIP", k1, k2)
					continue Loop2
				}
			}

			it1 := dp1.Iterator()
			it2 := dp2.Iterator()
			p := PairSim{
				Path1: sk.p1,
				Path2: sk.p2,
				Sim:   NewSimilarity(it1, it2),
			}
			if p.Sim.Identical() {
				seenIdent[sk] = p
			} else {
				seen[sk] = p
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

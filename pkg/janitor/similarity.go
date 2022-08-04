package janitor

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

type Similarity struct {
	BytesSame int64   // number of bytes corresponding to files that match
	BytesDiff int64   // number of bytes corresponding to files that don't match
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

func (s Similarity) EqualPathSim(s2 Similarity, delta float64) bool {
	return math.Abs(s.PathSim-s2.PathSim) < delta
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

// keys are paths within an implicit walkPath
func GetPairSims(all map[string]DirPrint, log io.Writer) []PairSim {
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
			if Child(k1, k2) || Child(k2, k1) {
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

			// ### PairSim eliding
			//
			// Given example identical dirs a/b and foo/b, there are a bunch of pairwise similarity comparisons that
			// we assume to be non-interesting, and should be dropped.
			// Below follows an example table of how we iterate over paths, form pairs, and which cases can be skipped.
			//
			// Note:
			// 1) we iterate the tree top down, in sorted order, thus reaching parents before their children, but that's only true for each individual loop,
			//    not their combination (since we have a dual loop)
			// 2) only pairs iterated after finding the identical case can be dropped during processing. the others need a post-processing loop.
			//
			// Example path pair    | relation to identical | codename | action & comments
			// k1=a      k2=foo     | parent and parent     | PP       | drop (in theory, could also happen to be full similarity, but that's not useful
			//                                                           for this exercise cause then these keys would take the roles that we currently give to a/b and foo/b.
			//                                                           in practice, thanks to the CC rule, the parents won't be equal if the children are, because
			//                                                           the children would have never been added if the parents were already known to be identical.
			// k1=a      k2=foo/b   | match and parent      | MP       | drop. some similarity, but known not equal.
			// k1=a      k2=foo/b/c | parent and child      | CP       | drop. some similarity, but known not equal.
			// k1=a/b    k2=foo     | match and parent      | MP       | drop. some similarity, but known not equal.
			// k1=a/b    k2=foo/b   | identical             |          | keep and report! this is the main one the user cares about!
			// k1=a/b    k2=foo/b/c | child and match       | CM       | drop. some similarity, but known not equal.
			// k1=a/b/c  k2=foo     | parent and child      | CP       | drop. some similarity, but known not equal.
			// k1=a/b/c  k2=foo/b   | child and match       | CM       | drop. some similarity, but known not equal.
			// k1=a/b/c  k2=foo/b/c | child and child       | CC       | drop. known equal.
			// k1=a/b/c  k2=foo/b/d | child and child       | CC       | drop. possibly, but likely not equal.
			//
			// The chosen approach to deal with this is in two steps:
			// 1) recognize as many cases as possible during the main iteration loop (because it's cheap to consider the small set of
			//    identical pairs found so far, for each iteration step) - these are the cases that surface after we found the identical pair.
			// 2) handle the other cases (the ones we found prior to knowing about the identical pair) in a subsequent processing loop.
			//
			// There are probably alternative solutions, like using an index to, upon finding identical pairs, cleaning up pairs that we created prior; or do the scanning depth-first (e.g. find similarities of all children before reporting on the parents), try to find the largest identical subtrees first, etc.  This might be the best solution, but needs more thought and brainpower which i don't have right now.
			// So for now, this should work well enough, and it's quite simple.
			//
			// Note that it's not clear-cut that all these cases are always non-interesting. For example, a PP case could still be interesting if they have content
			// that is similar, beyond their equal children. However, we assume the user would then clean up the identical children, at which point a subsequent run
			// would expose the remaining similarities of the parents.

			for ident := range seenIdent {
				// CC
				if BothChildren(ident.p1, ident.p2, k1, k2) {
					fmt.Fprintln(log, "IN-PROCESS DROP:", ident, "were identical. Skipping 2 children      ", k1, k2)
					continue Loop2
				}
				// CM
				if AChildAMatch(ident.p1, ident.p2, k1, k2) {
					fmt.Fprintln(log, "IN-PROCESS DROP:", ident, "were identical. Skipping child and match ", k1, k2)
					continue Loop2
				}
				// CP
				if AChildAParent(ident.p1, ident.p2, k1, k2) {
					fmt.Fprintln(log, "IN-PROCESS DROP:", ident, "were identical. Skipping child and parent", k1, k2)
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

			// there is nothing interesting about pairs that have no files in common
			if p.Sim.BytesSame == 0 {
				continue
			}

			pairSims = append(pairSims, p)
		}
	}

	// there's a way to do this without allocating a whole new slice. future optimization.
	filteredPairSims := make([]PairSim, 0, len(pairSims))

	// because we couldn't drop all irrelevant pairs during processing, do so now, in a dedicated loop
Loop1:
	for _, p := range pairSims {
		for ident := range seenIdent {
			// PP
			if BothChildren(p.Path1, p.Path2, ident.p1, ident.p2) {
				// our paths (the parents) should not be identical (otherwise the children would not have been added above), and thus can be dropped
				if p.Sim.Identical() {
					panic("this should never happen. post-process case PP found an identical pairsim of children and parents")
				}
				fmt.Fprintln(log, "POST-PROCESS DROP:", ident, "were identical. Skipping 2 parents       ", p.Path1, p.Path2)
				continue Loop1
			}
			// MP
			if AChildAMatch(p.Path1, p.Path2, ident.p1, ident.p2) {
				fmt.Fprintln(log, "POST-PROCESS DROP:", ident, "were identical. Skipping match and parent", p.Path1, p.Path2)
				continue Loop1
			}
			// CP
			if AChildAParent(ident.p1, ident.p2, p.Path1, p.Path2) {
				fmt.Fprintln(log, "POST-PROCESS DROP:", ident, "were identical. Skipping child and parent", p.Path1, p.Path2)
				continue Loop1
			}
		}
		filteredPairSims = append(filteredPairSims, p)
	}
	pairSims = filteredPairSims

	// sort pairSims by Similarity
	sort.Slice(pairSims, func(i, j int) bool {
		return pairSims[i].Sim.Less(pairSims[j].Sim)
	})
	return pairSims

}

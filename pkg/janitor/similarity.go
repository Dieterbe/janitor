package janitor

import (
	"bytes"
	"fmt"
	"path/filepath"
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
	if s.BytesDiff > 0 {
		return false
	}
	// this is... probably good enough?
	// TODO bump higher
	if s.PathSim < 0.8 {
		return false
	}
	return true
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

// dir is the directory that was walked to obtain all given dirprints
func GetPairSims(dir string, all map[string]DirPrint) []PairSim {
	type seenKey struct {
		p1 string
		p2 string
	}
	seen := make(map[seenKey]PairSim)
	var pairSims []PairSim

	keys := make([]string, 0, len(all))
	for k := range all {
		keys = append(keys, k)
	}

	// make sure parent directories come before children directories.
	// will allow us to skip over testing children if the parents are already identical
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
			// e.g. pointless to compare /foo/bar/baz to /foo/bar , it's obvious they will have some similarity, but not due to redundancy of data warranting cleanup.
			if SubPath(k1, k2) || SubPath(k2, k1) {
				continue
			}
			sk := seenKey{p1: k1, p2: k2}
			if k1 > k2 {
				sk = seenKey{p1: k2, p2: k1}
			}
			if _, ok := seen[sk]; ok {
				// this is a pretty naive way to avoid duplicates
				// would it be better to only iterate sub-ranges of the keys? that would require getting all keys first
				// for now, this is good enough
				continue
			}

			// try to skip these directories if we know their parents/grandparents to be identical already
			// in that case, there's no value in saying their children are also identical.
			parentSeenKey := sk
			// fmt.Println("### DIE seenkey", sk)
			for parentSeenKey.p1 != "/" && parentSeenKey.p2 != "/" {

				parentSeenKey.p1 = filepath.Dir(parentSeenKey.p1)
				parentSeenKey.p2 = filepath.Dir(parentSeenKey.p2)
				if len(parentSeenKey.p1) < len(dir) || len(parentSeenKey.p2) < len(dir) {
					// we have reached a pair that certainly won't have been considered
					// because one of the paths is not within the walked directory.
					break
				}

				parents, ok := seen[parentSeenKey]
				if !ok {
					// try the flipped order
					parentSeenKey.p1, parentSeenKey.p2 = parentSeenKey.p2, parentSeenKey.p1
					parents, ok = seen[parentSeenKey]
				}
				if ok {
					if parents.Sim.Identical() {
						// fmt.Println("DIE parent", parentSeenKey, "identical. CAN SKIP WOOHOOO")
						continue Loop2
					} else {
						// fmt.Println("DIE parent", parentSeenKey, "NOT IDENTICAL")
						// if the parents are certainly not identical, then
						// we can give up trying to find an identical grandparent.
						// instead let's compute the similarity for the children.
						break
					}
				}
				// if we haven't seen parents, than perhaps we've seen the grandparents.. probably not though (?)
				// fmt.Println("DIE parent", parentSeenKey, "NOT FONUD")
			}

			it1 := dp1.Iterator()
			it2 := dp2.Iterator()
			p := PairSim{
				Path1: sk.p1,
				Path2: sk.p2,
				Sim:   NewSimilarity(it1, it2),
			}
			seen[sk] = p
			pairSims = append(pairSims, p)
		}
	}

	// sort pairSims by Similarity
	sort.Slice(pairSims, func(i, j int) bool {
		return pairSims[i].Sim.Less(pairSims[j].Sim)
	})
	return pairSims

}

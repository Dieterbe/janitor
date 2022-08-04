package app

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Dieterbe/janitor/pkg/janitor"
	"github.com/google/go-cmp/cmp"
)

// arguably this file belongs in the janitor package (next to similarity.go), but then we wouldn't have access to app.Walk which we rely on (circular dependency)

// TestGetPairSims tests whether a given set of dirprints results in the expected pairsims
// specifically, due to 2 directories being identical, we expect a lot of other results being
// elided from the resultset. For more information, see the documentation for GetPairSims()
func TestGetPairSims(t *testing.T) {

	dpDir3 := janitor.DirPrint{
		Path: "dir3",
		Files: []janitor.FilePrint{
			mkFilePrint("c.txt", "c\n"),
		},
	}
	dpDir4 := janitor.DirPrint{
		Path: "dir4",
		Files: []janitor.FilePrint{
			mkFilePrint("d.txt", "d\n"),
		},
	}
	dpDir2 := janitor.DirPrint{
		Path: "dir2",
		Files: []janitor.FilePrint{
			mkFilePrint("b.txt", "b\n"),
		},
		Dirs: []janitor.DirPrint{
			dpDir3,
			dpDir4,
		},
	}
	dpDir1 := janitor.DirPrint{
		Path: "dir1",
		Files: []janitor.FilePrint{
			mkFilePrint("a", "a\n"),
			mkFilePrint("foo", "foo\n"),
		},
		Dirs: []janitor.DirPrint{
			dpDir2,
		},
	}
	dpDir2Zip := janitor.DirPrint{
		Path: "dir2.zip",
		Dirs: []janitor.DirPrint{
			dpDir2,
		},
	}
	dpDirRoot := janitor.DirPrint{
		Path: "fakedir",
		Dirs: []janitor.DirPrint{
			dpDir1,
			dpDir2Zip,
		},
	}

	all := map[string]janitor.DirPrint{
		".":                  dpDirRoot,
		"dir1":               dpDir1,
		"dir1/dir2":          dpDir2,
		"dir1/dir2/dir3":     dpDir3,
		"dir1/dir2/dir4":     dpDir4,
		"dir2.zip":           dpDir2Zip,
		"dir2.zip/dir2":      dpDir2,
		"dir2.zip/dir2/dir3": dpDir3,
		"dir2.zip/dir2/dir4": dpDir4,
	}

	// the hierachy is:
	// .
	// dir1/
	// dir1/a
	// dir1/foo
	// dir1/dir2/                      <----
	// dir1/dir2/b.txt                     |
	// dir1/dir2/dir3/c.txt                |
	// dir1/dir2/dir4/d.txt                |-- these two dirs are identical, and cause a lot of the other results to be elided...
	// dir2.zip                            |
	// dir2.zip/dir2/                  <----
	// dir2.zip/dir2/b.txt
	// dir2.zip/dir2/dir3/c.txt
	// dir2.zip/dir2/dir4/d.txt

	// ... and the rest of the dirs have nothing in common, so they are ommitted too.
	// this test would be a bit more useful if there were some other directories that do have files in common, so they would show up.
	// perhaps the test below will cover that case.

	expected := []janitor.PairSim{
		{
			Path1: "dir1/dir2",
			Path2: "dir2.zip/dir2",
			Sim: janitor.Similarity{
				BytesSame: 6,
				BytesDiff: 0,
				PathSim:   1,
			},
		},
	}

	if diff := cmp.Diff(expected, janitor.GetPairSims(all, os.Stderr)); diff != "" {
		t.Errorf("GetPairSims() mismatch (-want +got):\n%s", diff)
	}
}

// TestGetPairSimsTestdata tests whether the dirprints for the testdata directory results in the expected pairsims
// specifically, due to 2 directories being identical, we expect a lot of other results being
// elided from the resultset. For more information, see the documentation for GetPairSims()
func TestGetPairSimsTestdata(t *testing.T) {
	// get absolute directory for the testdata directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir, err = filepath.Abs(filepath.Join(dir, filepath.Join("..", "testdata")))
	if err != nil {
		t.Fatal(err)
	}
	f := os.DirFS(dir)
	_, all, err := WalkFS(f, dir, janitor.Sha256FingerPrint, ioutil.Discard)

	pairSims := janitor.GetPairSims(all, os.Stderr)
	expected := []janitor.PairSim{
		// non-identical copies of dir2: dir2-and-more and dir2-contents.zip are pitted against each other, and against all other copies
		// "dir1", "dir1/dir2", "dir1.zip", "dir1.zip/dir1", "dir1.zip/dir1/dir2", "dir2.zip", "dir2.zip/dir2",
		// the specific pathsim scores (and their differences) are not all that interesting here cause they are not that well calibrated.
		{
			Path1: "dir1",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 6,
				PathSim:   0.13043478260869568,
			},
		},
		{
			Path1: "dir1.zip/dir1",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 6,
				PathSim:   0.13043478260869568,
			},
		},
		{
			Path1: "dir1",
			Path2: "dir2-and-more",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 16,
				PathSim:   0.1578947368421053,
			},
		},
		{
			Path1: "dir1.zip/dir1",
			Path2: "dir2-and-more",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 16,
				PathSim:   0.1578947368421053,
			},
		},
		{
			Path1: "dir1.zip",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 6,
				PathSim:   0.16666666666666663,
			},
		},
		{
			Path1: "dir1.zip",
			Path2: "dir2-and-more",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 16,
				PathSim:   0.20833333333333337,
			},
		},
		{
			Path1: "dir1.zip/dir1/dir2",
			Path2: "dir2-and-more",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 10,
				PathSim:   0.21052631578947367,
			},
		},
		{
			Path1: "dir2-and-more",
			Path2: "dir2.zip/dir2",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 10,
				PathSim:   0.21052631578947367,
			},
		},
		{
			Path1: "dir1/dir2",
			Path2: "dir2-and-more",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 10,
				PathSim:   0.21052631578947367,
			},
		},
		{
			Path1: "dir2-and-more",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 10,
				PathSim:   0.21739130434782605,
			},
		},
		{
			Path1: "dir2-and-more",
			Path2: "dir2.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 10,
				PathSim:   0.5789473684210527,
			},
		},
		{
			Path1: "dir1.zip/dir1/dir2",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   0.17391304347826086,
			},
		},
		{
			Path1: "dir1/dir2",
			Path2: "dir2-contents.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   0.17391304347826086,
			},
		},
		{
			Path1: "dir2-contents.zip",
			Path2: "dir2.zip",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   0.17391304347826086,
			},
		},
		{
			Path1: "dir2-contents.zip",
			Path2: "dir2.zip/dir2",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   0.17391304347826086,
			},
		},

		// identical copies of dir2 and dir1
		// dir1 can be found in: dir1, dir1.zip/dir1
		// dir2 can be found in: dir1/dir2, dir1.zip/dir1/dir2, dir2.zip/dir2, which results in 3 pairs.
		// However, since the dir1-dir1.zip/dir1 relationship is already reported, the relationships of the dir2 dirs within them are elided.
		{
			Path1: "dir1/dir2",
			Path2: "dir2.zip/dir2",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   1,
			},
		},
		{
			Path1: "dir1.zip/dir1/dir2",
			Path2: "dir2.zip/dir2",
			Sim: janitor.Similarity{
				BytesSame: 2,
				BytesDiff: 0,
				PathSim:   1,
			},
		},
		{
			Path1: "dir1",
			Path2: "dir1.zip/dir1",
			Sim: janitor.Similarity{
				BytesSame: 8,
				BytesDiff: 0,
				PathSim:   1,
			},
		},
	}

	if diff := cmp.Diff(expected, pairSims); diff != "" {
		t.Errorf("GetPairSims() mismatch (-want +got):\n%s", diff)
	}

}

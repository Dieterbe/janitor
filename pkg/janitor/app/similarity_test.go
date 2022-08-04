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

func TestGetPairSimsTestdata(t *testing.T) {
	// TODO: after completing TestGetPairSims - which is a simplified version, we can improve this more and check that the set of the returned pairsims checks out.
	// e.g. for one thing,  since we have:
	/*
	    (janitor.PairSim) {
	    Path1: (string) (len=73) "dir1",
	    Path2: (string) (len=82) "dir1.zip/dir1",
	    Sim: (janitor.Similarity) <Similarity bytes=1.00 path=1.00>
	   },


	   this one should not show up:

	   (janitor.PairSim) {
	    Path1: (string) (len=73) "dir1",
	    Path2: (string) (len=87) "dir1.zip/dir1/dir2",
	    Sim: (janitor.Similarity) <Similarity bytes=0.25 path=0.27>
	   },
	*/
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

}

package app

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Dieterbe/janitor/pkg/janitor"
	"github.com/google/go-cmp/cmp"
)

// TestWalk tests whether a walk over an in-memory FS results in the expected DirPrints.
// TODO do we have a test anywhere that also checks for adding the "intermediate" dirprints?
// similar test that has a full path AND a zip file?
func TestWalk(t *testing.T) {

	var tests = []struct {
		name string
		data fstest.MapFS
		want janitor.DirPrint
		err  error
	}{
		{"main", janitor.DataMain, janitor.DataMainPrint, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dirPrint, _, err := WalkFS(tt.data, "/test/in-memory/"+tt.name+".zip", janitor.Sha256FingerPrint, os.Stderr)
			if err != tt.err {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.err)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, dirPrint); diff != "" {
				t.Errorf("Walk() mismatch (-want +got):\n%s", diff)
			}
		})

	}
}

func mkFilePrint(p string, content string) janitor.FilePrint {
	return janitor.FilePrint{
		Path: p,
		Size: int64(len(content)),
		Hash: sha256.Sum256([]byte(content)),
	}
}

// TestWalkTestData tests whether a walk over the sample testdata results in the expected DirPrints.
func TestWalkTestdata(t *testing.T) {
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
	root, all, err := WalkFS(f, dir, janitor.Sha256FingerPrint, ioutil.Discard)

	dpDir2 := janitor.DirPrint{
		Path: "dir2",
		Files: []janitor.FilePrint{
			mkFilePrint("b.txt", "b\n"),
		},
	}
	dpDir1 := janitor.DirPrint{
		Path: "dir1",
		Files: []janitor.FilePrint{
			mkFilePrint("a", "a\n"),
			mkFilePrint("foo", "foo\n"),
		},
		Dirs: []janitor.DirPrint{dpDir2},
	}
	dpDir2AndMore := janitor.DirPrint{
		Path: "dir2-and-more",
		Files: []janitor.FilePrint{
			mkFilePrint("b.txt", "b\n"),
			mkFilePrint("otherfile", "otherfile\n"),
		},
	}
	dpUnrelated := janitor.DirPrint{
		Path: "unrelated",
		Files: []janitor.FilePrint{
			mkFilePrint("unrelated.txt", "completely unrelated\n"),
		},
	}
	dpDir1Zip := janitor.DirPrint{
		Path: "dir1.zip",
		Dirs: []janitor.DirPrint{
			dpDir1,
		},
	}
	dpDir2Zip := janitor.DirPrint{
		Path: "dir2.zip",
		Dirs: []janitor.DirPrint{
			dpDir2,
		},
	}
	dpDir2ContentsZip := dpDir2
	dpDir2ContentsZip.Path = "dir2-contents.zip"

	dpDirRoot := janitor.DirPrint{
		Path: ".",
		Dirs: []janitor.DirPrint{
			dpDir1,
			dpDir1Zip,
			dpDir2AndMore,
			dpDir2ContentsZip,
			dpDir2Zip,
			dpUnrelated,
		},
	}

	expAll := map[string]janitor.DirPrint{
		".":                  dpDirRoot,
		"dir1":               dpDir1,
		"dir1/dir2":          dpDir2,
		"dir2-and-more":      dpDir2AndMore,
		"unrelated":          dpUnrelated,
		"dir1.zip":           dpDir1Zip,
		"dir1.zip/dir1":      dpDir1,
		"dir1.zip/dir1/dir2": dpDir2,
		"dir2-contents.zip":  dpDir2ContentsZip,
		"dir2.zip":           dpDir2Zip,
		"dir2.zip/dir2":      dpDir2,
	}
	if err != nil {
		t.Errorf("Walk() error = %v", err)
	}
	if diff := cmp.Diff(dpDirRoot, root); diff != "" {
		t.Errorf("Walk() root mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expAll, all); diff != "" {
		t.Errorf("Walk() all mismatch (-want +got):\n%s", diff)
	}
}

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

// TODO run same tests on "regular directory"? these are not specific to zip
// TODO do we have a test anywhere that also checks for adding the "intermediate" dirprints?
// similar test that has a full path AND a zip file?
func TestTraverse(t *testing.T) {

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

			dirPrint, _, err := WalkZip(tt.data, "test/in-memory/"+tt.name+".zip", janitor.Sha256FingerPrint, os.Stderr)
			if err != tt.err {
				t.Errorf("WalkZip() error = %v, wantErr %v", err, tt.err)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, dirPrint); diff != "" {
				t.Errorf("WalkZip() mismatch (-want +got):\n%s", diff)
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

func TestTraverseTestdata(t *testing.T) {
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
		Path: "testdata",
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
		dir:                                      dpDirRoot,
		filepath.Join(dir, "dir1"):               dpDir1,
		filepath.Join(dir, "dir1/dir2"):          dpDir2,
		filepath.Join(dir, "dir2-and-more"):      dpDir2AndMore,
		filepath.Join(dir, "unrelated"):          dpUnrelated,
		filepath.Join(dir, "dir1.zip"):           dpDir1Zip,
		filepath.Join(dir, "dir1.zip/dir1"):      dpDir1,
		filepath.Join(dir, "dir1.zip/dir1/dir2"): dpDir2,
		filepath.Join(dir, "dir2-contents.zip"):  dpDir2ContentsZip,
		filepath.Join(dir, "dir2.zip"):           dpDir2Zip,
		filepath.Join(dir, "dir2.zip/dir2"):      dpDir2,
	}
	if err != nil {
		t.Errorf("WalkFS() error = %v", err)
	}
	if diff := cmp.Diff(dpDirRoot, root); diff != "" {
		t.Errorf("WalkFS() root mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expAll, all); diff != "" {
		t.Errorf("WalkFS() all mismatch (-want +got):\n%s", diff)
	}
}

package app

import (
	"crypto/sha256"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	"github.com/google/go-cmp/cmp"
)

// TODO run same tests on "regular directory"? these are not specific to zip
// TODO do we have a test anywhere that also checks for adding the "intermediate" dirprints?
// similar test that has a full path AND a zip file?
func TestTraverse(t *testing.T) {

	var tests = []struct {
		name string
		data fstest.MapFS
		want hdc.DirPrint
		err  error
	}{
		{"main", hdc.DataMain, hdc.DataMainPrint, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dirPrint, _, err := WalkZip(tt.data, "test/in-memory/"+tt.name+".zip", hdc.Sha256FingerPrint, os.Stderr)
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

func mkFilePrint(p string, content string) hdc.FilePrint {
	return hdc.FilePrint{
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
	root, all, err := WalkFS(f, dir, hdc.Sha256FingerPrint, ioutil.Discard)

	dpDir2 := hdc.DirPrint{
		Path: "dir2",
		Files: []hdc.FilePrint{
			mkFilePrint("b.txt", "b\n"),
		},
	}
	dpDir1 := hdc.DirPrint{
		Path: "dir1",
		Files: []hdc.FilePrint{
			mkFilePrint("a", "a\n"),
			mkFilePrint("foo", "foo\n"),
		},
		Dirs: []hdc.DirPrint{dpDir2},
	}
	dpDir2AndMore := hdc.DirPrint{
		Path: "dir2-and-more",
		Files: []hdc.FilePrint{
			mkFilePrint("b.txt", "b\n"),
			mkFilePrint("otherfile", "otherfile\n"),
		},
	}
	dpUnrelated := hdc.DirPrint{
		Path: "unrelated",
		Files: []hdc.FilePrint{
			mkFilePrint("unrelated.txt", "completely unrelated\n"),
		},
	}
	dpDir1Zip := hdc.DirPrint{
		Path: "dir1.zip",
		Dirs: []hdc.DirPrint{
			dpDir1,
		},
	}
	dpDir2Zip := hdc.DirPrint{
		Path: "dir2.zip",
		Dirs: []hdc.DirPrint{
			dpDir2,
		},
	}
	dpDir2ContentsZip := dpDir2
	dpDir2ContentsZip.Path = "dir2-contents.zip"

	dpDirRoot := hdc.DirPrint{
		Path: "testdata",
		Dirs: []hdc.DirPrint{
			dpDir1,
			dpDir1Zip,
			dpDir2AndMore,
			dpDir2ContentsZip,
			dpDir2Zip,
			dpUnrelated,
		},
	}

	expAll := map[string]hdc.DirPrint{
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

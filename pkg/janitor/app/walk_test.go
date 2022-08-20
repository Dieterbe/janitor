package app

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Dieterbe/janitor/pkg/janitor"
	"github.com/Dieterbe/janitor/pkg/janitor/errfs"
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

// regular File: open, stat, read, close // DONE
// Dir: open, stat, read, close, readDir, direntryinfo // DONE
// zip File: open, stat, read, close // TODO
// same but regular file and dir inside of zipfile // TODO

// TestWalkWithErrors tests behavior when any of the FS, File or Directory operations fail, where the file is a regular file
// we can look at the case where the file is a zipfile, later
func TestWalkWithErrors(t *testing.T) {
	// Set up a structure with a possible failure on the file2 within a directory, amongst some other files.
	// This allows proper testing of "abort only the current directory" behavior
	baseFS := fstest.MapFS{
		"a":         {Data: []byte("foo")},
		"dir/file1": {Data: []byte("foo")},
		"dir/file2": {Data: []byte("bar")},
		"dir/file3": {Data: []byte("foobar")},
		"z":         {Data: []byte("bar")},
	}
	printsNoErr := janitor.DirPrint{
		Path: ".",
		Files: []janitor.FilePrint{
			{Path: "a", Size: 3, Hash: janitor.FooHash},
			{Path: "z", Size: 3, Hash: janitor.BarHash},
		},
		Dirs: []janitor.DirPrint{
			{
				Path: "dir",
				Files: []janitor.FilePrint{
					{Path: "file1", Size: 3, Hash: janitor.FooHash},
					{Path: "file2", Size: 3, Hash: janitor.BarHash},
					{Path: "file3", Size: 6, Hash: janitor.FooBarHash},
				},
			},
		},
	}
	printsDirSkipped := printsNoErr
	printsDirSkipped.Dirs = nil

	var tests = []struct {
		name   string
		baseFS fs.FS
		errors map[string]errfs.Errs
		want   janitor.DirPrint
		err    error
	}{
		{
			name:   "none",
			baseFS: baseFS,
			errors: nil,
			want:   printsNoErr,
			err:    nil,
		},
		{
			name:   "file-open",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir/file2": {
					Open: &fs.PathError{Op: "read", Path: "dir/file2", Err: fs.ErrNotExist},
				},
			},
			// if we can't open any file in a dir, we should skip the dir
			want: printsDirSkipped,
			err:  nil,
		},
		{
			name:   "file-stat",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir/file2": {
					Stat: errors.New("dummy stat error"),
				},
			},
			// Stat() is not used on regular files when walking, so no problem!
			want: printsNoErr,
			err:  nil,
		},
		{
			name:   "file-read",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir/file2": {
					Read: &fs.PathError{Op: "read", Path: "dir/file2", Err: fs.ErrPermission},
				},
			},
			// if we can't read any file in a dir, we should skip the dir
			want: printsDirSkipped,
			err:  nil,
		},
		{
			name:   "file-close",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir/file2": {
					Close: errors.New("dummy close error"),
				},
			},
			// Closing err after read only should not be a problem!
			want: printsNoErr,
			err:  nil,
		},
		{
			name:   "dir-open",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					Open: &fs.PathError{Op: "read", Path: "dir", Err: fs.ErrNotExist},
				},
			},
			// if we can't open a dir, we should skip it
			want: printsDirSkipped,
			err:  nil,
		},
		{
			name:   "dir-stat",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					Stat: errors.New("some stat error"),
				},
			},
			// per the fs.ReadDir docs, whether or not fs implements ReadDirFS, Stat() should never even be called while walking (ReadDir())
			// thus, no problem!
			want: printsNoErr,
			err:  nil,
		},
		{
			name:   "dir-read",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					Read: errors.New("some read error"),
				},
			},
			// there's no way we would ever call Read() on a dir, so no problem!
			want: printsNoErr,
			err:  nil,
		},
		{
			name:   "dir-close",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					Close: errors.New("dummy close error"),
				},
			},
			// dir.Close() will be called, but this should be treated as harmless after reading only.
			want: printsNoErr,
			err:  nil,
		},
		{
			name:   "dir-readDir",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					ReadDir: errors.New("some read error"),
				},
			},
			// if we can't readDir(), we have to skip the dir.
			want: printsDirSkipped,
			err:  nil,
		},
		{
			name:   "dir-entryinfo",
			baseFS: baseFS,
			errors: map[string]errfs.Errs{
				"dir": {
					DirEntryInfo: []error{nil, errors.New("some error")},
				},
			},
			// if we can't readDir(), we have to skip the dir.
			want: printsDirSkipped,
			err:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPrint, _, err := WalkFS(errfs.NewErrFS(tt.baseFS, tt.errors), "/test/in-memory/"+tt.name+".zip", janitor.Sha256FingerPrint, os.Stderr)
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

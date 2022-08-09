// Package errfs provides a fs.FS implementation with configurable errors, useful for testing.
package errfs

import (
	"io/fs"
)

var _ fs.FS = ErrFS{}
var _ fs.File = ErrFile{}
var _ fs.ReadDirFile = ErrDir{}
var _ fs.DirEntry = ErrDirEntry{}

type Errs struct {
	Open  error
	Stat  error
	Read  error
	Close error

	ReadDir      error   // errors for ReadDir() for directories.
	DirEntryInfo []error // errors for the fs.DirEntry's (which are returned by ReadDir()) Info() methods.
}

// ErrFS wraps fs.FS and returns ErrFile and ErrDir upon calling the Open method.
// The caller controls, on a per-path basis, whether the open fails, and which of the fs.File or fs.ReadDirFile methods should fail as well, or
// any of the DirEntry's returned by ReadDir().
// In other words, any method of ErrFS, or any of the things it returns (or the things they return), can be configured to fail.
type ErrFS struct {
	fs  fs.FS
	err map[string]Errs
}

func NewErrFS(f fs.FS, err map[string]Errs) fs.FS {
	return ErrFS{
		fs:  f,
		err: err,
	}
}

// Open fails as configured, or returns a fs.File (or fs.ReadDirFile) which will fail as configured.
func (efs ErrFS) Open(name string) (fs.File, error) {
	errs, ok := efs.err[name]
	if ok && errs.Open != nil {
		return nil, errs.Open
	}
	f, err := efs.fs.Open(name)
	if err != nil {
		return f, err
	}
	dir, ok := f.(fs.ReadDirFile)
	if ok {
		return ErrDir{
			errStat:         errs.Stat,
			errRead:         errs.Read,
			errReadDir:      errs.ReadDir,
			errClose:        errs.Close,
			errDirEntryInfo: errs.DirEntryInfo,
			f:               dir,
		}, nil
	}
	return ErrFile{
		errStat:  errs.Stat,
		errRead:  errs.Read,
		errClose: errs.Close,
		f:        f,
	}, nil
}

// ErrFile is like fs.File but it will fail any of its methods as configured
type ErrFile struct {
	errStat  error
	errRead  error
	errClose error
	f        fs.File
}

func (f ErrFile) Stat() (fs.FileInfo, error) {
	if f.errStat != nil {
		return nil, f.errStat
	}
	return f.f.Stat()
}
func (f ErrFile) Read(b []byte) (int, error) {
	if f.errRead != nil {
		return 0, f.errRead
	}
	return f.f.Read(b)
}
func (f ErrFile) Close() error {
	if f.errClose != nil {
		return f.errClose
	}
	return f.f.Close()
}

// ErrDir is like fs.ReadDirFile but it will fail any of its methods,
// or those of the returned fs.DirEntry's, as configured.
type ErrDir struct {
	errStat         error
	errRead         error
	errReadDir      error
	errClose        error
	errDirEntryInfo []error
	f               fs.ReadDirFile
}

func (f ErrDir) Stat() (fs.FileInfo, error) {
	if f.errStat != nil {
		return nil, f.errStat
	}
	return f.f.Stat()
}
func (f ErrDir) Read(b []byte) (int, error) {
	if f.errRead != nil {
		return 0, f.errRead
	}
	// note, calling Read() on a directory always fails AFAIK
	return f.f.Read(b)
}

func (f ErrDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.errReadDir != nil {
		return nil, f.errReadDir
	}
	entries, err := f.f.ReadDir(n)
	if err != nil {
		return entries, err
	}
	var dirEntries []fs.DirEntry
	dirEntries = make([]fs.DirEntry, 0, len(entries))
	for i, entry := range entries {
		var errInfo error
		if i < len(f.errDirEntryInfo) {
			errInfo = f.errDirEntryInfo[i]
		}
		dirEntries = append(dirEntries, ErrDirEntry{
			errInfo:  errInfo,
			DirEntry: entry,
		})
	}
	return dirEntries, nil
}

func (f ErrDir) Close() error {
	if f.errClose != nil {
		return f.errClose
	}
	return f.f.Close()
}

// ErrDirEntry is like fs.DirEntry but it will fail its Info() method as configured.
type ErrDirEntry struct {
	errInfo error
	fs.DirEntry
}

func (e ErrDirEntry) Info() (fs.FileInfo, error) {
	if e.errInfo != nil {
		return nil, e.errInfo
	}
	return e.DirEntry.Info()
}

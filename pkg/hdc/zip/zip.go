package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/fswalk"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
)

var errUnsupported = errors.New("unsupported feature")

func WalkZipFile(dir, base string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, error) {

	path := filepath.Join(dir, base)

	zipfs, err := zip.OpenReader(path)

	perr(err)
	defer zipfs.Close()

	// FYI. zipfs implements these types
	var _ fs.FS = zipfs
	var _ zip.ReadCloser = *zipfs
	var _ zip.Reader = zipfs.Reader

	return WalkZip(zipfs.Reader, dir, base, fpr, log)
}

func WalkZip(f zip.Reader, dir string, base string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, error) {
	logPrefix := "WalkZIP: " + filepath.Join(dir, base)
	fmt.Fprintln(log, "INF", logPrefix, "starting walk")
	var dirPrints []hdc.DirPrint

	// Note that WalkDir first processes a directory, then its children
	// For an example of a walking order, please see the README.md

	// p is the filename within the zip file, and d is the corresponding dirEntry
	// Note: it is assumed zip files on the system are trusted. Either way we won't actually extract them onto the system
	// (only in memory to get the hashes)
	// That said, we may want to add zip slip protection here. see https://gosamples.dev/unzip-file/
	walkDirFn := func(p string, d fs.DirEntry, err error) error {
		logPrefix := logPrefix + ": WalkDir " + p
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "callback received error", err, "..aborting")
			return err
		}

		info, err := d.Info()
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "d.info() error:", err, "..aborting")
			return err
		}

		if d.Name() == "__MACOSX" && info.IsDir() {
			fmt.Fprintln(log, "INF", logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}

		base := filepath.Base(p)

		if info.IsDir() {
			// entering a new directory. start our DirPrint to capture FilePrint's in this directory
			dirPrints = append(dirPrints, hdc.DirPrint{
				Path: base,
			})
			fmt.Fprintln(log, "INF", logPrefix, "PUSH: this is our current directory to add FilePrints into")
		} else {
			fmt.Fprintln(log, "INF", logPrefix, "fingerprinting...")
			fd, err := f.Open(p)
			perr(err)
			dirPrints[len(dirPrints)-1].Files = append(dirPrints[len(dirPrints)-1].Files, fpr(base, fd))
		}

		return nil
	}
	doneDirFn := func(p string, d fs.DirEntry) {
		logPrefix := logPrefix + ": DoneDir " + p
		// we are done with a directory, add it to its parent
		// unless this was the root directory, which has no parent and will be the ultimate DirPrint to return below
		if len(dirPrints) > 1 {
			fmt.Fprintln(log, "INF", logPrefix, "POP: adding this dir to its parent")
			popped := dirPrints[len(dirPrints)-1]
			dirPrints = dirPrints[:len(dirPrints)-1]
			dirPrints[len(dirPrints)-1].Dirs = append(dirPrints[len(dirPrints)-1].Dirs, popped)
			return
		}
		fmt.Fprintln(log, "INF", logPrefix, "POP: this dir is the root and is complete")
	}
	err := fswalk.WalkDir(&f, ".", walkDirFn, doneDirFn)
	if err != nil {
		return hdc.DirPrint{}, err
	}
	if len(dirPrints) != 1 {
		panic(fmt.Sprintf("unexpected number of dirPrints %d: %v", len(dirPrints), dirPrints))
	}
	return dirPrints[0], err

}

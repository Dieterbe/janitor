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

func FingerPrintFile(dir, base string, fp hdc.FingerPrinter, log io.Writer) {

	path := filepath.Join(dir, base)

	zipfs, err := zip.OpenReader(path)

	perr(err)
	defer zipfs.Close()

	// FYI. zipfs implements these types
	var _ fs.FS = zipfs
	var _ zip.ReadCloser = *zipfs
	var _ zip.Reader = zipfs.Reader

	FingerPrint(zipfs.Reader, dir, base, fp, log)
}

func FingerPrint(f zip.Reader, dir string, base string, fp hdc.FingerPrinter, log io.Writer) error {
	logPrefix := "unzip: " + filepath.Join(dir, base)
	fmt.Fprintln(log, "INF", logPrefix, "starting walk")
	var dirObjects []hdc.DirPrint

	// Note that WalkDir first processes a directory, then its children
	// For an example of a walking order, please see the README.md

	// p is the filename within the zip file, and d is the corresponding dirEntry
	// Note: it is assumed zip files on the system are trusted. Either way we won't actually extract them onto the system
	// (only in memory to get the hashes)
	// That said, we may want to add zip slip protection here. see https://gosamples.dev/unzip-file/
	walkDirFn := func(p string, d fs.DirEntry, err error) error {
		fmt.Fprintln(log, "DIE", logPrefix, "Walking fname within zip", p)
		logPrefix := logPrefix + ": WalkDir " + p
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "callback received error", err, "..aborting")
			return err
		}

		if d.Name() == "__MACOSX" {
			fmt.Fprintln(log, "INF", logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "error", err, "..skipping")
			return err
		}
		info, err := d.Info()
		if err != nil {
			fmt.Fprintln(log, "err", logPrefix, "d.info() error:", err, "..skipping")
			return err
		}

		if info.IsDir() {
			// entering a new directory. start our dirObject to capture file objects in this directory
			dirObjects = append(dirObjects, hdc.DirPrint{
				Path: p,
			})
			fmt.Fprintln(log, "INF", logPrefix, "PUSH adding dirobject with path", p)
		} else {
			fd, err := f.Open(p)
			perr(err)
			dirObjects[len(dirObjects)-1].Children = append(dirObjects[len(dirObjects)-1].Children, fp.Add(p, fd))
		}

		return nil
	}
	doneDirFn := func(p string, d fs.DirEntry) {
		fmt.Fprintln(log, "DIE", "POP done walking directory", p)
		// we are done with a directory, add it to its parent
		// unless this was the root directory, which has no parent and will be the ultimate object to return below
		if len(dirObjects) > 1 {
			popped := dirObjects[len(dirObjects)-1]
			dirObjects = dirObjects[:len(dirObjects)-1]
			dirObjects[len(dirObjects)-1].Children = append(dirObjects[len(dirObjects)-1].Children, popped)
			dirObjects = dirObjects[:len(dirObjects)-1]
		}
	}
	err := fswalk.WalkDir(&f, ".", walkDirFn, doneDirFn)
	//finishAssimilationsMaybe("", true)
	// sum last dir
	return err

}

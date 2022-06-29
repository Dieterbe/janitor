package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

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

	type assimilation struct {
		dir string
	}
	var assimilations []assimilation
	var objects []hdc.Print
	// Note that WalkDir first processes a directory, then its children
	// For an example of a walking order, please see the README.md

	// p is the filename within the zip file, and d is the corresponding dirEntry
	// Note: it is assumed zip files on the system are trusted. Either way we won't actually extract them onto the system
	// (only in memory to get the hashes)
	// That said, we may want to add zip slip protection here. see https://gosamples.dev/unzip-file/
	err := fs.WalkDir(&f, ".", func(p string, d fs.DirEntry, err error) error {
		fmt.Fprintln(log, "DIE", logPrefix, "Walking fname within zip", p)
		logPrefix := logPrefix + ": WalkDir " + p
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "callback received error", err, "..aborting")
			return err
		}
		if d.Name() == "." {
			return nil
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

		// if inDir != "" && inDir != dirname {
		// 	// we finished traversing all the children of the last directory we walked. we must assimilate.
		// 	inDir = ""
		// }
		if info.IsDir() {
			// entering a new directory. wrap up any previous assimilation(s) as needed, and start a new one.

			// 1) unwind the stack of assimilations to finish off the previous ones, they have a dirName that is not a prefix our current path.
			for i := len(assimilations) - 1; i >= 0; i-- {
				if !strings.HasPrefix(p, assimilations[i].dir) {
					fmt.Fprintln(log, "INF", logPrefix, "unwinding assimilation", assimilations[i].dir)
					assimilations = assimilations[:i]
				}
			}

			// 2) create our new assimilation

			assimilations = append(assimilations, assimilation{dir: p})
		} else {
			dirname := filepath.Dir(p)
			// if this file is not contained within the dir, we need to assimilate the dir first
			fd, err := f.Open(p)
			perr(err)
			objects = append(objects, fp.Add(p, fd))
		}

		return nil
	})
	// sum last dir
	return err

}

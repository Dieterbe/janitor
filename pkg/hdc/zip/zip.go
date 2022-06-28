package zip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
)

var errUnsupported = errors.New("unsupported feature")

func FingerPrintFile(dir, base string, fp hdc.FingerPrinter) {

	path := filepath.Join(dir, base)
	fmt.Println("ZIP: IdentifyFile:", path)

	zipfs, err := zip.OpenReader(path)

	perr(err)
	defer zipfs.Close()

	// FYI. zipfs implements these types
	var _ fs.FS = zipfs
	var _ zip.ReadCloser = *zipfs
	var _ zip.Reader = zipfs.Reader

	FingerPrint(zipfs.Reader, dir, base, fp)
}

func FingerPrint(f zip.Reader, dir string, base string, fp hdc.FingerPrinter) error {
	logPrefix := "unzip: " + filepath.Join(dir, base)
	//fmt.Println(logPrefix, "start walk")

	// p is the filename within the zip file, and d is the corresponding dirEntry
	// Note: it is assumed zip files on the system are trusted. Either way we won't actually extract them onto the system
	// (only in memory to get the hashes)
	// That said, we may want to add zip slip protection here. see https://gosamples.dev/unzip-file/
	err := fs.WalkDir(&f, ".", func(p string, d fs.DirEntry, err error) error {
		logPrefix := logPrefix + ": WalkDir " + p
		//fmt.Println(logPrefix)
		if err != nil {
			fmt.Println(logPrefix, "callback received error", err, "..aborting")
			return err
		}
		if d.Name() == "." {
			//	fmt.Println(logPrefix, "ignoring .")
			return nil
		}
		if d.Name() == "__MACOSX" {
			fmt.Println(logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}
		if err != nil {
			fmt.Println(logPrefix, "error", err, "..skipping")
			return err
		}
		info, err := d.Info()
		if err != nil {
			fmt.Println(logPrefix, "d.Info() error:", err, "..skipping")
			return err
		}
		if info.IsDir() {
			fmt.Println(logPrefix, p, "is a dir. which is not supported, skipping")
			return errUnsupported
		}
		fd, err := f.Open(p)
		perr(err)

		fp.Add(p, fd)

		return nil
	})
	return err

}

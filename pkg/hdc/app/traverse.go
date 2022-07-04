package app

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/fswalk"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
)

func canonicalPath(dir, p string) (string, error) {
	// TODO: also make this follow symlinks?
	// the idea would be that if multiple paths point to the same file/dir,
	// we recognize it somehow and deduplicate it in the output, or at least only scan once.

	absPath := filepath.Join(dir, p)

	// filepath.Abs is a bit poorly named IMHO. what happens here is we give it an absolute path,
	// and it returns the "canonical" path with things like ./ and /../ cleaned up
	return filepath.Abs(absPath)
}

func walkZipFile(dir, base string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, map[string]hdc.DirPrint, error) {

	path := filepath.Join(dir, base)

	zipfs, err := zip.OpenReader(path)

	perr(err)
	defer zipfs.Close()

	// FYI. zipfs implements these types
	var _ fs.FS = zipfs
	var _ zip.ReadCloser = *zipfs
	var _ zip.Reader = zipfs.Reader

	return WalkZip(zipfs, path, fpr, log)
}

func WalkZip(f fs.FS, walkPath string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, map[string]hdc.DirPrint, error) {
	return Walk(f, "WalkZIP: ", walkPath, fpr, log)
}

func WalkFS(f fs.FS, walkPath string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, map[string]hdc.DirPrint, error) {
	return Walk(f, "WalkFS : ", walkPath, fpr, log)
}

// WalkFS walks the filesystem rooted at walkPath (could be a directory or a zip file)
// and generates the Prints for all folders, files and zip files encountered
// it returns the root DirPrint as all individual dirprints by path
func Walk(f fs.FS, prefix, walkPath string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, map[string]hdc.DirPrint, error) {
	logPrefix := prefix + walkPath
	fmt.Fprintln(log, "INF", logPrefix, "starting walk")
	var dirPrints []hdc.DirPrint                  // stack of dirprints in progress during walking
	allDirPrints := make(map[string]hdc.DirPrint) // all dirprints encountered.

	// Note that WalkDir first processes a directory, then its children
	// For an example of a walking order, please see the README.md

	// p is the filename within the zip file, and d is the corresponding dirEntry
	// Note: it is assumed zip files on the system are trusted. Either way we won't actually extract them onto the system
	// (only in memory to get the hashes)
	// That said, we may want to add zip slip protection here. see https://gosamples.dev/unzip-file/

	walkDirFn := func(p string, d fs.DirEntry, err error) error {
		logPrefix := logPrefix + ": WalkDir " + p
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "callback received error", err, "..skipping") // zip aborting
			return fs.SkipDir
		}

		canPath, err := canonicalPath(walkPath, p)

		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "failed to get canonical path for this entry", err, "..skipping")
			return fs.SkipDir
		}

		logPrefix = prefix + ": WalkDir " + canPath

		info, err := d.Info()
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "d.info() error:", err, "..skipping") // zip abort
			return fs.SkipDir
		}

		if d.Name() == "__MACOSX" && info.IsDir() {
			fmt.Fprintln(log, "INF", logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}

		base := filepath.Base(canPath)

		if info.IsDir() {
			// entering a new directory. start our DirPrint to capture FilePrint's in this directory
			dirPrints = append(dirPrints, hdc.DirPrint{
				Path: base,
			})
			fmt.Fprintln(log, "INF", logPrefix, "PUSH: this is our current directory to add FilePrints into")
		} else {
			if filepath.Ext(p) == ".zip" {
				fp, ok := allDirPrints[canPath]
				if ok {
					fmt.Fprintf(log, "INF %s fingerprint for this zip already exists. (original path %q, this path %q). skipping", logPrefix, fp.Path, p)
					return nil
				}
				fmt.Fprintln(log, "INF", logPrefix, "fingerprinting as an zip directory...")
				fp.Path = canPath // TODO is it useful to include the dir for all these? it's implied / can be deduplicated
				fp, all, err := walkZipFile(walkPath, p, fpr, log)
				if err != nil {
					return err
				}
				for k, v := range all {
					allDirPrints[k] = v
				}
				allDirPrints[canPath] = fp
				dirPrints[len(dirPrints)-1].Dirs = append(dirPrints[len(dirPrints)-1].Dirs, fp)
			} else {
				fmt.Fprintln(log, "INF", logPrefix, "fingerprinting as standalone file...")
				fd, err := f.Open(p)
				perr(err)
				dirPrints[len(dirPrints)-1].Files = append(dirPrints[len(dirPrints)-1].Files, fpr(base, fd))
			}
		}

		return nil

	}
	doneDirFn := func(p string, d fs.DirEntry) error {
		canPath, err := canonicalPath(walkPath, p)

		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "failed to get canonical path for this entry", err, "..skipping")
			panic("this should never happen. WalkDirFunc should have already returned an error")
			//return fs.SkipDir
		}

		logPrefix := logPrefix + ": DoneDir " + p

		allDirPrints[canPath] = dirPrints[len(dirPrints)-1] // our stack should always have at least 1 element.

		// we are done with a directory, add it to its parent
		// unless this was the root directory, which has no parent and will be the ultimate DirPrint to return below
		if len(dirPrints) > 1 {
			fmt.Fprintln(log, "INF", logPrefix, "POP: adding this dir to its parent")
			popped := dirPrints[len(dirPrints)-1]
			dirPrints = dirPrints[:len(dirPrints)-1]
			dirPrints[len(dirPrints)-1].Dirs = append(dirPrints[len(dirPrints)-1].Dirs, popped)
			return nil
		}
		fmt.Fprintln(log, "INF", logPrefix, "POP: this dir is the root and is complete")
		return nil
	}
	err := fswalk.WalkDir(f, ".", walkDirFn, doneDirFn)
	if err != nil {
		return hdc.DirPrint{}, nil, err
	}
	if len(dirPrints) != 1 {
		panic(fmt.Sprintf("unexpected number of dirPrints %d: %v", len(dirPrints), dirPrints))
	}
	return dirPrints[0], allDirPrints, nil
}

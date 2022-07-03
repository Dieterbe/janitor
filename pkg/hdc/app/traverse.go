package app

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/fswalk"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc/zip"
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

// WalkFS walks the filesystem rooted at dir and generates the Prints for all folders, files and zip files encountered
// it returns the root DirPrint as all individual dirprints by path
func WalkFS(f fs.FS, dir string, fpr hdc.FingerPrinter, log io.Writer) (hdc.DirPrint, map[string]hdc.DirPrint, error) {
	logPrefix := "WalkFS : " + dir
	fmt.Fprintln(log, "INF", logPrefix, "starting walk")
	var dirPrints []hdc.DirPrint                  // stack of dirprints in progress during walking
	allDirPrints := make(map[string]hdc.DirPrint) // all dirprints encountered.

	walkDirFn := func(p string, d fs.DirEntry, err error) error {
		logPrefix := logPrefix + ": WalkDir " + p
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "callback received error", err, "..skipping")
			return fs.SkipDir
		}

		canPath, err := canonicalPath(dir, p)

		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "failed to get canonical path for this entry", err, "..skipping")
			return fs.SkipDir
		}

		logPrefix = "WalkFS : WalkDir " + canPath

		info, err := d.Info()
		if err != nil {
			fmt.Fprintln(log, "ERR", logPrefix, "d.info() error:", err, "..skipping")
			return fs.SkipDir
		}

		if d.Name() == "__MACOSX" && info.IsDir() {
			fmt.Fprintln(log, "INF", logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}

		if info.IsDir() {
			// entering a new directory. start our DirPrint to capture FilePrint's in this directory
			dirPrints = append(dirPrints, hdc.DirPrint{
				Path: p,
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
				zip.WalkZipFile(dir, p, fpr, log)
				allDirPrints[canPath] = fp
				//m.allDirPaths = append(m.allDirPaths, canPath) // not sure yet if useful
			} else {
				fmt.Fprintln(log, "INF", logPrefix, "fingerprinting as standalone file...")
				fd, err := f.Open(p)
				perr(err)
				dirPrints[len(dirPrints)-1].Files = append(dirPrints[len(dirPrints)-1].Files, fpr(p, fd))
			}
		}

		return nil

	}
	doneDirFn := func(p string, d fs.DirEntry) error {
		canPath, err := canonicalPath(dir, p)

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

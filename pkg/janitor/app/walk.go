package app

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Dieterbe/fswalk"
	"github.com/Dieterbe/janitor/pkg/janitor"
)

func walkZipReader(fd fs.File, path string, fpr janitor.FingerPrinter, log io.Writer) (janitor.DirPrint, map[string]janitor.DirPrint, error) {

	// fd is an io.Reader, but we need an io.ReaderAt; so "convert" it
	var buf bytes.Buffer
	size, err := io.Copy(&buf, fd)
	if err != nil {
		return janitor.DirPrint{}, nil, err
	}

	readerAt := bytes.NewReader(buf.Bytes())

	zipfs, err := zip.NewReader(readerAt, size)
	if err != nil {
		return janitor.DirPrint{}, nil, err
	}

	return WalkZip(zipfs, path, fpr, log)
}

func WalkZip(f fs.FS, walkPath string, fpr janitor.FingerPrinter, log io.Writer) (janitor.DirPrint, map[string]janitor.DirPrint, error) {
	return Walk(f, "WalkZIP: ", walkPath, fpr, log, true)
}

func WalkFS(f fs.FS, walkPath string, fpr janitor.FingerPrinter, log io.Writer) (janitor.DirPrint, map[string]janitor.DirPrint, error) {
	return Walk(f, "WalkFS : ", walkPath, fpr, log, false)
}

// WalkFS walks the filesystem rooted at walkPath (absolute path to a directory or zip file)
// and generates the Prints for all folders, files and zip files encountered
// it returns the root DirPrint and all individual dirprints by path within walkPath (which is implicit)
// crit means whether any error should fail the entire walk at the root level, or only skip the directory where the error occurs
func Walk(f fs.FS, prefix, walkPath string, fpr janitor.FingerPrinter, log io.Writer, crit bool) (janitor.DirPrint, map[string]janitor.DirPrint, error) {
	if !strings.HasPrefix(walkPath, "/") {
		panic(fmt.Sprintf("expected an absolute path. not %q - may not be strictly necessary, but it makes output clearer. this should never happen", walkPath))
	}
	logPrefix := prefix + walkPath
	fmt.Fprintln(log, "INF", logPrefix+": START!!")
	var dpStack []janitor.DirPrint                // dirprints in progress during walking.
	var dpAll = make(map[string]janitor.DirPrint) // to be returned

	// Note that WalkDir first processes a directory, then its children

	// p is the filename within the zip file (or walked dir), and d is the corresponding dirEntry
	// Note that we never extract zip files onto the filesystem.  (only in memory to get the hashes)
	// Thus, zip slip protection as in https://gosamples.dev/unzip-file/ is not needed.
	// in fact, for this tool, let's deliberately allow path elements like ../../foo/bar, because eliding them would remove information about the path within the zip.

	walkDirFn := func(p string, d fs.DirEntry, err error) error {

		logPrefix := logPrefix + ": WalkDir " + p

		// handleErr logs the error, and for a critical error, reports the failure, otherwise skips
		handleErr := func(msg string, err error) error {
			if !crit {
				fmt.Fprintln(log, "WARN", logPrefix, msg, err, "..skipping dir")
				return fs.SkipDir
			}
			fmt.Fprintln(log, "ERR", logPrefix, msg, err, "..aborting")
			return err
		}

		if err != nil {
			return handleErr("received Stat(root) or ReadDir(dir) error", err)
		}

		info, err := d.Info()
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintln(log, "WARN", logPrefix, "file disappeared while walking. error:", err, "..skipping this file")
				return nil
			}
			return handleErr("d.info() error", err)
		}

		if d.Name() == "__MACOSX" && info.IsDir() {
			fmt.Fprintln(log, "INF", logPrefix, "don't descend into this one, it's not real important data")
			return fs.SkipDir
		}

		if info.IsDir() {
			// entering a new directory. start our DirPrint to capture FilePrint's in this directory
			dpStack = append(dpStack, janitor.DirPrint{Path: filepath.Base(p)})
			fmt.Fprintln(log, "INF", logPrefix, "PUSH: this is our current directory to add FilePrints into")
		} else {
			if filepath.Ext(p) == ".zip" {
				fmt.Fprintln(log, "INF", logPrefix, "fingerprinting as a zip directory...")
				fd, err := f.Open(p)
				if err != nil {
					return handleErr("f.Open() error", err)
				}
				path := filepath.Join(walkPath, p)
				dp, all, err := walkZipReader(fd, path, fpr, log)
				if err != nil {
					return handleErr("walkZip returned error:", err)
				}
				fd.Close() // ignore error. AFAIK this is fine after read-only access
				for k, v := range all {
					// normally if you call a walk function, the paths of returned dirprints don't include the walkPath prefix, as it is implied.
					// since we called walk within our walk, we have to prepend the portion of the path after (within) *our* walkPath
					dpAll[filepath.Join(p, k)] = v
				}
				// normally if you call a walk function, dp.Path is "." for the root dir (or in this case, the zip file), as the path is implied from the walkpath.
				// since we called within our walk, we must set path properly (which is per definition always the basename)
				dp.Path = filepath.Base(p)
				dpAll[p] = dp
				dpStack[len(dpStack)-1].Dirs = append(dpStack[len(dpStack)-1].Dirs, dp)
			} else {
				fmt.Fprintln(log, "INF", logPrefix, "fingerprinting as standalone file...")
				fd, err := f.Open(p)
				if err != nil {
					return handleErr("f.Open() error", err)
				}
				pr, err := fpr(filepath.Base(p), fd)
				if err != nil {
					return handleErr("Fingerprint (io.Read) returned error:", err)
				}
				err = fd.Close()
				if err != nil {
					fmt.Fprintln(log, "WARN", logPrefix, "fd.Close() returned error:", err, "..afaik these are harmless after read-only access. so ignoring")
				}
				dpStack[len(dpStack)-1].Files = append(dpStack[len(dpStack)-1].Files, pr)
			}
		}

		return nil
	}

	doneDirFn := func(p string, d fs.DirEntry, err error) error {
		logPrefix := logPrefix + ": DoneDir " + p

		if err != nil {
			// walking this dir was aborted
			fmt.Fprintln(log, "INF", logPrefix, "POP: discarding directory due to error")
			dpStack = dpStack[:len(dpStack)-1]
			return nil
		}

		dpAll[p] = dpStack[len(dpStack)-1] // our stack should always have at least 1 element.

		// we are done with a directory, add it to its parent
		// unless this was the root directory, which has no parent and will be the ultimate DirPrint to return below
		if len(dpStack) > 1 {
			fmt.Fprintln(log, "INF", logPrefix, "POP: adding this dir to its parent")
			popped := dpStack[len(dpStack)-1]
			dpStack = dpStack[:len(dpStack)-1]
			dpStack[len(dpStack)-1].Dirs = append(dpStack[len(dpStack)-1].Dirs, popped)
			return nil
		}
		fmt.Fprintln(log, "INF", logPrefix, "POP: this dir is the root and is complete")
		return nil
	}
	err := fswalk.WalkDir(f, ".", walkDirFn, doneDirFn)
	if err != nil {
		return janitor.DirPrint{}, nil, err
	}
	if len(dpStack) != 1 {
		panic(fmt.Sprintf("unexpected number of dirPrints %d: %v", len(dpStack), dpStack))
	}
	return dpStack[0], dpAll, nil
}

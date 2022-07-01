package app

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/fswalk"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc/zip"
)

// traverse walks the filesystem rooted at dir, which is provided only for printing
func traverse(f fs.FS, dir string, m *model) {
	walkDirFn := func(p string, d fs.DirEntry, err error) error {
		fmt.Fprintln(m.log, "INF WALKING", p)
		if p == "zipfiles" {
			fmt.Fprintln(m.log, "INF SKIPPING more zipfiles UNTIL THE PROGRAM IS MORE READY")
			return fs.SkipDir
		}
		if err != nil {
			fmt.Fprintln(m.log, "ERR failed to walk", p, err, "..skipping")
			return fs.SkipDir
		}

		// filepath.Abs is a bit poorly named IMHO. what happens here is we give it an absolute path,
		// and it returns the canonical path with things like ./ and /../ cleaned up
		absPath := filepath.Join(dir, p)
		canPath, err := filepath.Abs(absPath)
		if err != nil {
			fmt.Fprintln(m.log, "ERR failed to get canonical path for", p, err, "..skipping")
			return fs.SkipDir
		}
		if filepath.Ext(p) == ".zip" {
			fp, ok := m.objectData[canPath]
			if ok {
				fmt.Fprintf(m.log, "INF already have object for canonical Path %q (original path %q), skipping for path %q which resolves to same canonical path\n", canPath, fp.Path, p)
				return nil
			}
			fp.Path = absPath
			fmt.Fprintln(m.log, "INF STARTING ZIPFILE WALK FOR", p)
			zip.FingerPrintFile(dir, p, &fp.fp, m.log)
			fmt.Fprintln(m.log, "INF FINISHED ZIPFILE WALK FOR", p)
			m.objectData[canPath] = fp
			m.objectList = append(m.objectList, canPath)
		}
		return nil
	}
	doneDirFn := func(p string, d fs.DirEntry) {
		fmt.Fprintln(m.log, "INF DONE WALKING DIR", p, d.Name())
	}
	err := fswalk.WalkDir(f, ".", walkDirFn, doneDirFn)
	if err != nil {
		fmt.Fprintln(m.log, "ERR failed to walk", dir, err)
	}
}

type Object struct {
	Path string // the original, non-canonicalized absolute path.
	fp   hdc.Sha256FingerPrinter
}

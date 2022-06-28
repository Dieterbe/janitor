package app

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc/zip"
)

// traverse walks the filesystem rooted at dir, which is provided only for printing
func traverse(f fs.FS, dir string, m *model) {
	fs.WalkDir(f, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println("failed to walk", p, err, "..skipping")
			return fs.SkipDir
		}
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Println("failed to get absolute path for", p, err, "..skipping")
			return fs.SkipDir
		}
		if filepath.Ext(p) == ".zip" {
			fp, ok := m.objectData[absPath]
			if ok {
				fmt.Printf("already have object for absolute Path %q (relative path %q), skipping for relative path %q which resolves to same absolute path\n", absPath, fp.RelPath, p)
				return nil
			}
			fmt.Println("ADDING!")
			zip.FingerPrintFile(dir, p, &fp.fp)
			m.objectData[absPath] = fp
			m.objectList = append(m.objectList, absPath)
		}
		return nil
	})
}

type Object struct {
	RelPath string
	fp      hdc.Sha256FingerPrinter
}

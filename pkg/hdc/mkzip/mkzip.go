// package mkzip aids with making zip files
package mkzip

import (
	"archive/zip"
	"bytes"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
)

// note: we could perhaps use fstest.MapFS but then we lose ordering
func Do(files []hdc.Entry) ([]byte, zip.Reader) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		f, err := w.Create(file.Path)
		perr(err)
		_, err = f.Write([]byte(file.Body))
		perr(err)
	}

	err := w.Close()
	perr(err)
	b := buf.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	perr(err)
	return b, *zr
}

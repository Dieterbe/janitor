// package mkzip aids with making zip files
package mkzip

import (
	"archive/zip"
	"bytes"
)

type Entry struct {
	Path string
	Body string
}

func Do(files []Entry) ([]byte, zip.Reader, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		f, err := w.Create(file.Path)
		if err != nil {
			return nil, zip.Reader{}, err
		}
		_, err = f.Write([]byte(file.Body))
		if err != nil {
			return nil, zip.Reader{}, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, zip.Reader{}, err
	}
	b := buf.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, zip.Reader{}, err
	}
	return b, *zr, nil
}

func MustDo(files []Entry) ([]byte, zip.Reader) {
	b, r, err := Do(files)
	if err != nil {
		panic(err)
	}
	return b, r
}

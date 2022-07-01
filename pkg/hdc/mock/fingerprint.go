package mock

import (
	"io"
	"io/ioutil"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
)

func perr(err error) {
	if err != nil {
		panic(err)
	}
}

// MockFingerPrinter is a FingerPrinter that returns dummy fingerprints but tracks which files were added
type MockFingerPrinter struct {
	Entries []hdc.Entry
}

func (m *MockFingerPrinter) Add(path string, r io.Reader) hdc.FilePrint {
	body, err := ioutil.ReadAll(r)
	perr(err)
	entry := hdc.Entry{Path: path, Body: string(body)}
	m.Entries = append(m.Entries, entry)
	return hdc.FilePrint{
		Path: path,
	}
}

func (m *MockFingerPrinter) Reset() {
	m.Entries = m.Entries[:0]
}

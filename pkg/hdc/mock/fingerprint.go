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

type MockFingerPrinter struct {
	Entries []hdc.Entry
}

func (m *MockFingerPrinter) Add(path string, r io.Reader) {
	body, err := ioutil.ReadAll(r)
	perr(err)
	m.Entries = append(m.Entries, hdc.Entry{Path: path, Body: string(body)})
}

func (m *MockFingerPrinter) Reset() {
	m.Entries = m.Entries[:0]
}

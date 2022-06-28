package hdc

import (
	"crypto/sha256"
	"io"
)

type FingerPrinter interface {
	Add(path string, r io.Reader)
}

// ignored: owner, group, mode, modTime, etc
type Sha256FingerPrinter struct {
	Content      [32]byte // xor of the content of files within the zip
	ContentNamed [32]byte // xor of all paths in the zip and their content
}

// Add updates the fingerprint's content and contNamed hashes based on the given path and reader which signify a file
func (p *Sha256FingerPrinter) Add(path string, r io.Reader) {

	// add the checksum of the paths.
	h := sha256.New()
	_, err := h.Write([]byte(path))
	perr(err)

	sum := h.Sum(nil)
	for i := range sum {
		p.ContentNamed[i] ^= sum[i]
	}

	// add the checksum of the content.
	h.Reset()
	_, err = io.Copy(h, r)
	perr(err)

	sum = h.Sum(nil)
	for i := range sum {
		p.Content[i] ^= sum[i]
		p.ContentNamed[i] ^= sum[i]
	}
}

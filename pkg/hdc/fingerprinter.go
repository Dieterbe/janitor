package hdc

import (
	"crypto/sha256"
	"io"
)

type FingerPrinter interface {
	Add(path string, r io.Reader)
}

// Print represents a file's path and content.
// ignored: owner, group, mode, modTime, etc
type Print struct {
	Path string
	Hash [32]byte
}

type Sha256FingerPrinter struct {
	Prints []Print
}

// Add adds a fingerprint for the given file content
func (p *Sha256FingerPrinter) Add(path string, r io.Reader) {
	pr := Print{Path: path}

	h := sha256.New()
	_, err := io.Copy(h, r)
	perr(err)
	sum := h.Sum(nil)
	copy(pr.Hash[:], sum)

	p.Prints = append(p.Prints, pr)
}

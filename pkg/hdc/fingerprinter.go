package hdc

import (
	"crypto/sha256"
	"io"
)

type FingerPrinter interface {
	Add(path string, r io.Reader) FilePrint
}

// TODO: as we build all the state while traversing, this should probably be only the stateless hashing

type Object interface {
	Iterator() ObjectIterator
}

type ObjectIterator interface {
	Next() bool
	Value() FilePrint
}

// Print represents a file's path and content.
// ignored: owner, group, mode, modTime, etc
type FilePrint struct {
	Path string
	Hash [32]byte
}

func (fp FilePrint) Iterator() ObjectIterator {
	return &FilePrintIterator{
		read: false,
		fp:   fp,
	}
}

type FilePrintIterator struct {
	read bool
	fp   FilePrint
}

func (fpi *FilePrintIterator) Next() bool {
	read := fpi.read
	fpi.read = true
	return !read
}

func (fpi *FilePrintIterator) Value() FilePrint {
	return fpi.fp
}

type DirPrint struct {
	Path     string
	Children []Object
}

// merges
// - an iterator for all its file objects
// - an iterator for each child directory
func (dp DirPrint) Iterator() ObjectIterator {
	return &DirPrintIterator{
		read: false,
		dp:   dp,
	}
}

type DirPrintIterator struct {
	read bool
	dp   DirPrint
}

// TODO merge all dirprints into one
func (fpi *DirPrintIterator) Next() bool {
	read := fpi.read
	fpi.read = true
	return !read
}

func (fpi *DirPrintIterator) Value() FilePrint {
	return FilePrint{}
	//return fpi.dp
}

type Sha256FingerPrinter struct {
	Prints []FilePrint
}

// Add adds a fingerprint for the given file content
func (p *Sha256FingerPrinter) Add(path string, r io.Reader) FilePrint {
	pr := FilePrint{Path: path}

	h := sha256.New()
	_, err := io.Copy(h, r)
	perr(err)
	sum := h.Sum(nil)
	copy(pr.Hash[:], sum)

	p.Prints = append(p.Prints, pr)
	return pr
}

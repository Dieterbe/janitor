package janitor

import (
	"crypto/sha256"
	"fmt"
	"io"
)

// Print represents a file's path, size and content.
// ignored: owner, group, mode, modTime, etc
type FilePrint struct {
	Path string
	Size int64
	Hash [32]byte
}

func (fp FilePrint) String() string {
	return fmt.Sprintf("FilePrint %10d %x %s", fp.Size, fp.Hash, fp.Path)
}

type FingerPrinter func(path string, r io.Reader) FilePrint

// Sha256FingerPrint computes the sha256 based fingerprint for the given file content
func Sha256FingerPrint(path string, r io.Reader) FilePrint {
	pr := FilePrint{Path: path}

	h := sha256.New()
	var err error
	pr.Size, err = io.Copy(h, r)
	perr(err)
	sum := h.Sum(nil)
	copy(pr.Hash[:], sum)

	return pr
}

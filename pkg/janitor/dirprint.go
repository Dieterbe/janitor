package janitor

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
)

type DirPrint struct {
	Path  string // always the basename, or "." for the root dir
	Files []FilePrint
	Dirs  []DirPrint
}

func (dp DirPrint) String() string {
	return dp.string("")
}
func (dp DirPrint) string(indent string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%sDirPrint path: %q\n", indent, dp.Path)
	fmt.Fprintf(&buf, "%s  Files:\n", indent)
	for _, f := range dp.Files {
		buf.WriteString(indent + "     " + f.String() + "\n")
	}
	fmt.Fprintf(&buf, "%s  Dirs:\n", indent)
	for _, d := range dp.Dirs {
		indent += "    "
		buf.WriteString(d.string(indent) + "\n")
	}
	return buf.String()
}

func (dp DirPrint) Iterator() Iterator {

	// initialize all iterators and load up their first values (if any)
	var dpi DirPrintIterator
	dpi.path = dp.Path

	it := newFilePrintIterator(dp.Files)
	it.Next()
	dpi.its = append(dpi.its, it)
	dpi.itPaths = append(dpi.itPaths, "")

	for _, d := range dp.Dirs {
		it := d.Iterator()
		it.Next()
		dpi.its = append(dpi.its, it)
		dpi.itPaths = append(dpi.itPaths, d.Path)
	}

	return &dpi
}

type Iterator interface {
	Next() bool
	Value() (FilePrint, bool)
}

type FilePrintIterator struct {
	files []FilePrint
	idx   int
}

func newFilePrintIterator(files []FilePrint) Iterator {

	// sort all FilePrints by Hash. Note that this will change sorting of the original array
	sort.Slice(files, func(i, j int) bool {
		return bytes.Compare(files[i].Hash[:], files[j].Hash[:]) < 0
	})

	fpi := FilePrintIterator{
		files: files,
		idx:   -1,
	}

	return &fpi
}

func (fpi *FilePrintIterator) Next() bool {
	fpi.idx++
	return fpi.idx < len(fpi.files)
}

func (fpi *FilePrintIterator) Value() (FilePrint, bool) {
	if fpi.idx >= len(fpi.files) {
		return FilePrint{}, false
	}
	return fpi.files[fpi.idx], true
}

// DirPrintIterator is an iterator over a DirPrint. It merges:
// - an iterator for all its filePrints
// - an iterator for each child directory
type DirPrintIterator struct {
	path    string
	v       FilePrint
	valid   bool
	its     []Iterator // iterators over . and all subdirs
	itPaths []string   // paths describing each iterator ("" for current dir)
}

func (dpi *DirPrintIterator) Next() bool {
	var toAdvance int
	dpi.valid = false

	// collect the lowest value hash from all iterators' current values
	// this will be our return value
	for i, it := range dpi.its {
		v, ok := it.Value()
		if !ok {
			continue
		}
		if !dpi.valid {
			dpi.v = v
			dpi.valid = true
			toAdvance = i
		} else if bytes.Compare(v.Hash[:], dpi.v.Hash[:]) < 0 {
			dpi.v = v
			toAdvance = i
		}
	}

	if dpi.valid {

		// advance the iterator we have chosen to consume from
		dpi.its[toAdvance].Next()

		// when we pull a fileprint into its parent dir, we need to update the path accordingly.
		dpi.v.Path = filepath.Join(dpi.itPaths[toAdvance], dpi.v.Path)
	}

	return dpi.valid
}

func (fpi *DirPrintIterator) Value() (FilePrint, bool) {
	return fpi.v, fpi.valid
}

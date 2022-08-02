// Package testdata is a collection of data used in unit tests
// the interesting i'm trying to do here, is show how data evolves between different api's in different packages,
// (e.g. from zip to dirprint, from dirprint to iterator output, from iterator output to similarity calculation)
// Ideally I'ld like the output of one step to be the input of the next, so we can see how the data evolves.
// However, since different tests have different requirements and areas of focus, accommodating them all has
// ramifications on the entire sequence of structures which becomes hard to reason about.
// So instead, we sometimes switch to a different structure for a new test, but the similarities should be
// obvious.
// Tip: use "go to references" on the data to see where the data is being used.
// TODO: find a way to unexport these. i assume some of this needs to be imported from other packages.
// we also can't move this out into a 'testdata' package because it uses the janitor primitive types.
package janitor

import (
	"encoding/hex"
	"testing/fstest"
)

// various useful hashes for testing purposes.
var FooHash, BarHash, FooBarHash [32]byte
var h1, h2, h3, h4, h5, h6, h7 [32]byte
var DataMain fstest.MapFS
var DataMainPrint, DataMain2Print DirPrint
var DataMainIterated, DataMain2Iterated, DataMain3Iterated []FilePrint

func init() {
	// echo -n foo | sha256sum
	// 2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae  -
	// echo -n bar | sha256sum
	// fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9  -
	// echo -n foobar | sha256sum
	// c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2  -
	fooSlice, err := hex.DecodeString("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	perr(err)
	barSlice, err := hex.DecodeString("fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	perr(err)
	fooBarSlice, err := hex.DecodeString("c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2")
	perr(err)

	copy(FooHash[:], fooSlice)
	copy(BarHash[:], barSlice)
	copy(FooBarHash[:], fooBarSlice)

	// 1-5 are in order for one test
	h1S, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
	perr(err)
	h2S, err := hex.DecodeString("2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae")
	perr(err)
	h3S, err := hex.DecodeString("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	perr(err)
	h4S, err := hex.DecodeString("fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9")
	perr(err)
	h5S, err := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	perr(err)

	// 6,7 are alternative hashes that sort between h3 and h5, for another test
	h6S, err := hex.DecodeString("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	perr(err)
	h7S, err := hex.DecodeString("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	perr(err)

	copy(h1[:], h1S)
	copy(h2[:], h2S)
	copy(h3[:], h3S)
	copy(h4[:], h4S)
	copy(h5[:], h5S)
	copy(h6[:], h6S)
	copy(h7[:], h7S)

	// For future reference, with a script like below you can find 16
	// inputs that result in 16 hashes that will be in order.
	// Note that the inputs all have similar size, so for tests where you
	// want distinct size values this needs a tweak to introduce variability.
	// However, for now, this is overkill.
	// ~ ❯❯❯ cat hashes.sh
	// for patt in 00 11 22 33 44 55 66 77 88 99 aa bb cc dd ee ff; do
	// 	i=0
	// 	while true; do
	// 		echo -n $i | sha256sum | grep -q "^$patt" && echo $i $(echo -n $i | sha256sum) && break
	// 		((i++))
	// 	done
	// done

	// Data structures used as inputs/outputs of tests.

	// Note that iteration via fs.Walk always uses lexical ordering, so we don't need to worry about
	// ordering here.
	// we deliberately also have a few files with same content (and thus same hash).
	DataMain = fstest.MapFS{
		"bar/__MACOSX/somefile":     {Data: []byte("ignore this entry")},
		"bar/somefile":              {Data: []byte("foo")},
		"foo/bar/foobar.png.txt":    {Data: []byte("foobar")},
		"foo/bar/__MACOSX/another":  {Data: []byte("ignore this entry as well")},
		"foo/bar/__MACOSX/somefile": {},
		"foo/bar/somefile":          {Data: []byte("bar")},
		"foo/__MACOSX/somefile":     {},
		"foo/somefile":              {Data: []byte("foo")},
		"__MACOSX/somefile":         {},
		"somefile":                  {Data: []byte("foo")},
	}

	DataMainPrint = DirPrint{
		Path: ".",
		Files: []FilePrint{
			{
				Path: "somefile",
				Size: 3,
				Hash: FooHash,
			},
		},
		Dirs: []DirPrint{
			{
				Path: "bar",
				Files: []FilePrint{
					{
						Path: "somefile",
						Size: 3,
						Hash: FooHash,
					},
				},
			},
			{
				Path: "foo",
				Files: []FilePrint{
					{
						Path: "somefile",
						Size: 3,
						Hash: FooHash,
					},
				},
				Dirs: []DirPrint{
					{
						Path: "bar",
						Files: []FilePrint{
							{
								Path: "foobar.png.txt",
								Size: 6,
								Hash: FooBarHash,
							},
							{
								Path: "somefile",
								Size: 3,
								Hash: BarHash,
							},
						},
					},
				},
			},
		},
	}

	// Note the similarities with the DataMainPrint structure above.
	// note that the expected prints in the iterator output are ordered by their hash,
	// which is different from the order of the DirPrint structure, hence we here
	// deliberately try a few random orders between files in the same dir, which also
	// explains why we create a new structure here.
	DataMain2Print = DirPrint{
		Path: ".",
		Files: []FilePrint{
			{
				Path: "a",
				Size: 122,
				Hash: h2,
			},
			{
				Path: "z",
				Hash: h1,
				Size: 100,
			},
		},
		Dirs: []DirPrint{
			{
				Path: "b",
				Files: []FilePrint{
					{
						Path: "1",
						Hash: h4,
						Size: 444,
					},
					{
						Path: "2",
						Hash: h3,
						Size: 333,
					},
					{
						Path: "3",
						Hash: h5,
						Size: 555,
					},
				},
				Dirs: nil,
			},
		},
	}

	// we test: hierachy flattening, path concat, hash based ordering
	DataMain2Iterated = []FilePrint{
		{
			Path: "z",
			Hash: h1,
			Size: 100,
		},
		{
			Path: "a",
			Hash: h2,
			Size: 122,
		},
		{
			Path: "b/2",
			Hash: h3,
			Size: 333,
		},
		{
			Path: "b/1",
			Hash: h4,
			Size: 444,
		},
		{
			Path: "b/3",
			Hash: h5,
			Size: 555,
		},
	}

	// this is a fake iteration output of a supposedly 3rd DirPrint
	// it is defined by being similar, but different compared to DataMain2Iterated above.
	// this will help with similarity computation testing.
	DataMain3Iterated = []FilePrint{
		// in DataMain2Iterated but not here
		//{
		//	Path: "z",
		//	Hash: h1,
		//	Size: 100,
		//},
		// identical
		{
			Path: "a",
			Hash: h2,
			Size: 122,
		},
		// identical
		{
			Path: "b/2",
			Hash: h3,
			Size: 333,
		},
		// hash mismatch. note that this file contributes two times to the bytes differing, due to it existing with 2 different hashes.
		// whether this is correct behavior, is debatable. but anyway at least it's simple code.
		{
			Path: "b/1",
			Hash: h6, // need a hash here that sorts after h3
			Size: 444,
		},
		// file that we have, but DataMain2Iterated doesn't.
		{
			Path: "b/4",
			Hash: h7, // needs to sort after the hash above, but before h5
			Size: 7777,
		},
		// identical file to DataMain2Iterated but different path
		{
			Path: "3.copy", // similarity is 0.45
			Hash: h5,
			Size: 555,
		},
	}
}

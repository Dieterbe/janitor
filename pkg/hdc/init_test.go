package hdc

import "encoding/hex"

// some "random" hashes for use in testing
var h1, h2, h3, h4, h5, h6, h7 [32]byte
var dpToIterate DirPrint
var dpIterated []FilePrint

func init() {
	var err error
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

	dpToIterate = DirPrint{
		Path: "root",
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

	dpIterated = []FilePrint{
		{
			Path: "root/z",
			Hash: h1,
			Size: 100,
		},
		{
			Path: "root/a",
			Hash: h2,
			Size: 122,
		},
		{
			Path: "root/b/2",
			Hash: h3,
			Size: 333,
		},
		{
			Path: "root/b/1",
			Hash: h4,
			Size: 444,
		},
		{
			Path: "root/b/3",
			Hash: h5,
			Size: 555,
		},
	}
}

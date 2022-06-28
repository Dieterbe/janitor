package hdc

// Entry abstracts a file. It ignores ownership, mode, and timestamps.
type Entry struct {
	Path string
	Body string
}

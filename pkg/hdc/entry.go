package hdc

// Entry abstracts a file. It ignores ownership, mode, and timestamps.
// TODO used for what? zip specific?
type Entry struct {
	Path string
	Body string
}

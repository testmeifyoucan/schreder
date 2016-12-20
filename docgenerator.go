package schreder

// ITaggable is an interface that can tell doc generator
// that some test provides a tag. Usable for swagger documentation
// where tags help to group API endpoints
type ITaggable interface {
	Tag() string
}

// IDocGenerator describes a generator of documentation
// uses test suite as a source of information about API endpoints
type IDocGenerator interface {
	Generate(tests []Test) ([]byte, error)
}

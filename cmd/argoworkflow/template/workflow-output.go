package template

// OutputObject represents a output object
type OutputObject struct {
	Kind     OutputObjectKind
	File     string
	FileName string
	Value    string
}

// OutputObjectKind represents the type of the outout object.
// This is a type alias of string.
type OutputObjectKind string

const (
	// FileOutput represetnts a file path
	FileOutput OutputObjectKind = "file"
	// ValueOutput represents a string value
	ValueOutput OutputObjectKind = "string"
	// MarkdownOutput represents a markdown format string value
	MarkdownOutput OutputObjectKind = "markdown"
)

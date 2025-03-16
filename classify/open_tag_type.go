package classify

// OpenTagType represents the type of opening sequence for a tag in a Mustache
// template. It distinguishes between standard delimiters (default: '{{') and
// triple braces ('{{{').
type OpenTagType int8

const (
	// InvalidOpenTag represents an invalid or unrecognized opening tag sequence.
	InvalidOpenTag OpenTagType = iota

	// TripleBraceOpen represents the triple-brace opening sequence '{{{' for
	// unescaped variables.
	TripleBraceOpen

	// DelimiterOpen represents the standard opening delimiter (default: '{{') for
	// typical tags.
	DelimiterOpen
)

// String returns a human-readable representation of OpenTagType for error
// messages and test output.
func (ott OpenTagType) String() string {
	switch ott {
	case TripleBraceOpen:
		return "TripleBraceOpen"
	case DelimiterOpen:
		return "DelimiterOpen"
	case InvalidOpenTag:
		fallthrough
	default:
		return "InvalidOpenTag"
	}
}

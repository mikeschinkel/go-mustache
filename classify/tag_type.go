package classify

// TagTypes is a collection of TagType values.
type TagTypes []TagType

// TagType represents the semantic meaning of a tag in a Mustache template. Each
// tag type occupies the upper 4 bits in a Segment's bitmap representation.
type TagType uint8 // Fits into 4 bits

// IsValid determines if a TagType represents a semantically valid tag type.
// NotApplicable is the zero value for TagType which indicates that a Segment
// does not have one of the tag types defined by the Mustache spec.
func (tt TagType) IsValid() (valid bool) {
	return tt != NotApplicable
}

const (
	// NotApplicable represents a tag type that is not applicable. Not applicable is
	// used with segment types TextContent and Whitespace.
	NotApplicable TagType = iota

	// AmpersandUnescaped represents an unescaped variable tag using ampersand
	// syntax: {{&tag}}
	AmpersandUnescaped

	// BlockTag represents a block definition tag with dollar syntax:
	// {{$block}}...{{/block}}
	BlockTag

	// CommentTag represents a comment tag with exclamation point syntax:
	// {{! Comment goes here}}
	CommentTag

	// SetDelimiterTag represents a delimiter setting tag that uses the equals
	// character, e.g.: {{=<% %>=}}
	SetDelimiterTag

	// DotTag represents a current context tag with dot syntax: {{.}}
	DotTag

	// InvertedSectionTag represents an inverted section tag with caret syntax:
	// {{^section}}...{{/section}}
	InvertedSectionTag

	// ParentTag represents a parent template tag for inheritance:
	// {{<parent}}...{{/parent}}
	ParentTag

	// PartialTag represents a partial inclusion tag with greater-than syntax:
	// {{>partial}}
	PartialTag

	// VarTag represents a standard variable tag: {{var}}
	VarTag

	// SectionTag represents a section tag with hash syntax:
	// {{#section}}...{{/section}}
	SectionTag

	// TripleBraceUnescaped represents an unescaped variable tag with triple brace
	// syntax: {{{tag}}}
	TripleBraceUnescaped
)

// String returns a human-readable representation of TagType for error messages
// and test output. NotApplicable returns an empty strings to support ignoring
// in test output.
func (tt TagType) String() string {
	switch tt {
	case AmpersandUnescaped:
		return "AmpersandUnescaped"
	case BlockTag:
		return "BlockTag"
	case CommentTag:
		return "CommentTag"
	case SetDelimiterTag:
		return "SetDelimiterTag"
	case DotTag:
		return "DotTag"
	case InvertedSectionTag:
		return "InvertedSectionTag"
	case ParentTag:
		return "ParentTag"
	case PartialTag:
		return "PartialTag"
	case VarTag:
		return "VarTag"
	case SectionTag:
		return "SectionTag"
	case TripleBraceUnescaped:
		return "TripleBraceUnescaped"
	case NotApplicable:
		return ""
	default:
		return "UnknownTagType"
	}
}

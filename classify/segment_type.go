package classify

// SegmentTypes is a collection of SegmentType values.
type SegmentTypes []SegmentType

// SegmentType represents the structural role of a segment within a Mustache template.
// Each segment occupies the lower 4 bits in a Segment's bitmap representation.
type SegmentType uint8 // Fits into 4 bits

const (
	// CompleteTag represents a complete tag that does not span multiple lines.
	CompleteTag SegmentType = iota

	// BeginTag represents the beginning of an enclosing tag. Enclosing tags include
	// Section, InvertedSection, Block, and Parent tags.
	BeginTag

	// TagIdentifier represents a the alphanumeric identifier within a tag.
	TagIdentifier

	// EndTag represents the end of an enclosing tag (e.g., closing a section).
	EndTag

	// MultilineBegin represents the first line of a multi-line tag (e.g., comments).
	MultilineBegin

	// MultilineMiddle represents a middle line(s) in a multi-line tag.
	MultilineMiddle

	// MultilineEnd represents the last line in a multi-line tag.
	MultilineEnd

	// TextContent represents plain text content (not a valid "tag").
	TextContent

	// Whitespace represents text that consists only of inline whitespace meaning
	// spaces and tabs; also not a valid "tag".
	Whitespace

	// InvalidSegmentType represents an uninitialized or invalid segment type.
	InvalidSegmentType
)

// String returns a human-readable representation of SegmentType for debugging and logging.
func (t SegmentType) String() string {
	switch t {
	case InvalidSegmentType:
		return "InvalidSegmentType"
	case CompleteTag:
		return "CompleteTag"
	case TextContent:
		return "TextContent"
	case MultilineBegin:
		return "MultilineBegin"
	case MultilineMiddle:
		return "MultilineMiddle"
	case MultilineEnd:
		return "MultilineEnd"
	case Whitespace:
		return "Whitespace"
	case BeginTag:
		return "BeginTag"
	case TagIdentifier:
		return "TagIdentifier"
	case EndTag:
		return "EndTag"
	default:
		return "UnknownSegmentType"
	}
}

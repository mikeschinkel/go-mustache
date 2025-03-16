package classifier

type SegmentTypes []SegmentType
type SegmentType uint8 // Fits into 4 bits

const (
	CompleteTag SegmentType = iota
	BeginTag
	TagIdentifier
	EndTag
	MultilineBegin  // First line of a multi-line tag construct
	MultilineMiddle // Middle line of a multi-line tag construct
	MultilineEnd    // Last line of a multi-line tag construct
	TextContent     // Not a tag but a TagType to represent text
	Whitespace      // Text that is just whitespace, e.g. '\t' or ' '
	InvalidSegmentType
	IgnoredSegmentType
)

// String returns a human-readable representation of a segment type for debugging
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
	case IgnoredSegmentType:
		return ""
	default:
		return "UnknownSegmentType"
	}
}

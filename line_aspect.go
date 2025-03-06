package mustache

// LineAspect represents the role of a line in a Mustache template
type LineAspect int
type LineAspects = []LineAspect

const (
	InvalidLineType      LineAspect = iota // Non-initialized line type
	EmptyLine                              // Empty line type
	TextLine                               // Regular content, not a tag
	WhitespaceLine                         // Whitespace only content, not a tag
	InlineTagLine                          // Line with tag(s) mixed with other content
	StandaloneTagLine                      // Line with a single standalone tag
	MultilineTagBegin                      // First line of a multi-line tag construct
	MultilineTagMiddle                     // Middle line of a multi-line tag construct
	MultilineTagEnd                        // Last line of a multi-line tag construct
	UnescapedTagLine                       // Generic
	TripleBraceUnescaped                   // Unescaped tag w/this syntax: {{{tag}}}
	AmpersandUnescaped                     // Unescaped tag w/this syntax: {{&tag}}
)

// String returns a human-readable representation of LineAspect for debugging
func (lt LineAspect) String() string {
	switch lt {
	case InvalidLineType:
		return "InvalidTagLine"
	case EmptyLine:
		return "EmptyLine"
	case WhitespaceLine:
		return "WhitespaceLine"
	case TextLine:
		return "TextLine"
	case InlineTagLine:
		return "InlineTagLine"
	case StandaloneTagLine:
		return "StandaloneTagLine"
	case MultilineTagBegin:
		return "MultilineTagBegin"
	case MultilineTagMiddle:
		return "MultilineTagMiddle"
	case MultilineTagEnd:
		return "MultilineTagEnd"
	case UnescapedTagLine:
		return "UnescapedTagLine"
	case TripleBraceUnescaped:
		return "TripleBraceUnescaped"
	case AmpersandUnescaped:
		return "AmpersandUnescaped"
	default:
		return "UnknownLineType"
	}
}

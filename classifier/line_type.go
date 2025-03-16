package classifier

import (
	"reflect"
)

// LineTypes is a collection of LineType values.
type LineTypes []LineType

// LineType represents the classification of a line in a Mustache template.
// This classification determines how the line should be rendered.
type LineType uint8

const (
	// InvalidLineType represents a non-initialized line type.
	InvalidLineType LineType = iota

	// EmptyLine represents a line with no content (zero length).
	EmptyLine

	// InlineLine represents a line with tags mixed with other content.
	InlineLine

	// StandaloneLine represents a line with a single standalone tag.
	// Standalone tags are handled specially in Mustache (e.g., partials, sections, etc.).
	StandaloneLine

	// TextLine represents a line containing only text content, i.e. no tags and more
	// than just whitespace.
	TextLine

	// WhitespaceLine represents a line containing only whitespace.
	WhitespaceLine

	// NotApplicableLine is used when a line type is required but not applicable such
	// as when segment types are TextContent or Whitespace.
	NotApplicableLine
)

// Equal compares two LineTypes collections for equality. This is used when
// testing.
func (ts LineTypes) Equal(types LineTypes) bool {
	return reflect.DeepEqual(ts, types)
}

// String returns a human-readable representation of LineType for use in error
// messages.
func (t LineType) String() string {
	switch t {
	case EmptyLine:
		return "EmptyLine"
	case InlineLine:
		return "InlineLine"
	case StandaloneLine:
		return "StandaloneLine"
	case TextLine:
		return "TextLine"
	case WhitespaceLine:
		return "WhitespaceLine"
	case NotApplicableLine:
		return "NotApplicableLine"
	case InvalidLineType:
		return "InvalidLineType"
	default:
		return "UnknownLineType"
	}
}

package classifier

import (
	"reflect"
)

type LineTypes []LineType
type LineType int

const (
	InvalidLineType   LineType = iota // Non-initialized line type
	EmptyLine                         // Empty line type
	InlineLine                        // Line with tag(s) mixed with other content
	StandaloneLine                    // Line with a single standalone tag
	TextLine                          // Regular content, not a tag
	WhitespaceLine                    // Whitespace only content, not a tag
	NotApplicableLine                 // To be used when required but explicitly not applicable
)

func (ts LineTypes) Equal(types LineTypes) bool {
	return reflect.DeepEqual(ts, types)
}

// String returns a human-readable representation of Tag for debugging
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

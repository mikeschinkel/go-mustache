package classifier

import (
	"fmt"
	"reflect"
	"strings"
)

// Lines represents a collection of Line objects that make up a template.
type Lines []Line

// String returns a human-readable representation of Lines for debugging and logging.
func (ll Lines) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, line := range ll {
		sb.WriteString(line.String())
		if i == len(ll)-1 {
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteByte(']')
	return sb.String()
}

// Normalize cleans up Lines by setting empty Segments slices to nil.
// This allows deep equality comparisons for testing.
func (ll Lines) Normalize() Lines {
	for i := range ll {
		if len(ll[i].Segments) == 0 && ll[i].Segments != nil {
			ll[i].Segments = nil
		}
	}
	return ll
}

// Equal compares two Lines collections for deep equality.
// It normalizes both collections before comparison.
func (ll Lines) Equal(lines Lines) bool {
	if len(ll) != len(lines) {
		return false
	}
	return reflect.DeepEqual(ll.Normalize(), lines.Normalize())
}

// Line represents a single line in a Mustache template with its classification
// and the segments it contains.
type Line struct {
	Type     LineType
	Segments Segments
}

// String returns a human-readable representation of Line for debugging and logging.
// Simple line types (TextLine, WhitespaceLine) return just the type name.
// Complex line types include segment information.
func (l *Line) String() (str string) {
	typ := l.Type
	switch typ {
	case TextLine, WhitespaceLine:
		str = typ.String()
		goto end
	default:
	}
	str = fmt.Sprintf("{%s, %s}", typ.String(), l.Segments.String())
end:
	return str
}

// PastEOL determines if a position is past the end of the line.
func (l *Line) PastEOL(pos *int, len int) bool {
	return *pos >= len
}

// AtEOL determines if a position is exactly at the end of the line.
func (l *Line) AtEOL(pos *int, len int) bool {
	return *pos == len
}

package classifier

import (
	"fmt"
	"strings"
)

// InitialSegmentsCapacity defines the initial capacity for new Segments slices.
// This is just a micro-optimization to keep from having to churn memory for any
// lines having three or fewer segments because the author guesses that the vast
// majority of lines will have three or fewer segments.
const InitialSegmentsCapacity = 3

// Segments represents a collection of Segment values that make up a line in a
// template.
type Segments []Segment

// NewSegments creates a new empty Segments slice with default initial capacity.
func NewSegments() Segments {
	return make(Segments, 0, InitialSegmentsCapacity)
}

// Segment is a bitmap representation that combines SegmentType and TagType. The
// lower 4 bits represent the SegmentType, and the upper 4 bits represent the
// TagType.
type Segment uint8

// String returns a human-readable representation of Segment for error messages
// and test output.
//
// - TextContent segments return just "TextContent"
// - CompleteTag segments return the tag type string
// - Other segments return a combined string showing both segment and tag types
func (s Segment) String() (str string) {
	var tt TagType
	t := s.Type()
	if t == TextContent {
		str = "TextContent"
		goto end
	}
	tt = s.TagType()
	if t == CompleteTag {
		str = tt.String()
		goto end
	}
	str = fmt.Sprintf("{%s, %s}", t.String(), tt.String())
end:
	return str
}

// IsValid checks if all segments in the collection are valid. Returns false if
// any segment is invalid or if the collection is empty. Individual segments are
// invalid if they have an InvalidSegmentType, or if SegmentType is CompleteTag
// AND TagType is NotApplicable.
func (ss *Segments) IsValid() (valid bool) {
	if len(*ss) == 0 {
		goto end
	}
	for _, segment := range *ss {
		if segment.IsValid() {
			continue
		}
		goto end
	}
	valid = true
end:
	return valid
}

// String returns a human-readable representation of Segments for error messages
// and test output. For a single segment, it returns just that segment's string
// representation. For multiple segments, it returns a formatted list with
// indices.
func (ss *Segments) String() string {
	var sb strings.Builder
	if len(*ss) == 1 {
		sb.WriteString((*ss)[0].String())
		goto end
	}
	sb.WriteString("Segments: [")
	for i, segment := range *ss {
		sb.WriteString(fmt.Sprintf("[%d]", i))
		sb.WriteString(segment.String())
		if i == len(*ss)-1 {
			break
		}
		sb.WriteString(", ")
	}
	sb.WriteByte(']')
end:
	return sb.String()
}

// IsValid checks if a segment has a valid combination of segment type and tag
// type. Returns false if the segment type is InvalidSegmentType or if it's a
// CompleteTag with NotApplicable tag type, which is an invalid combination.
func (s Segment) IsValid() (valid bool) {
	// Check if low 4 bits are InvalidSegmentType
	if uint8(s&0x0F) == uint8(InvalidSegmentType) {
		goto end
	}

	// Check if segment is CompleteTag (low 4 bits) with NotApplicable (high 4 bits)
	if uint8(s&0x0F) == uint8(CompleteTag) && uint8(s>>4) == uint8(NotApplicable) {
		goto end
	}

	valid = true
end:
	return valid
}

// MakeSegment creates a bitmapped value combining SegmentType and TagType. The
// Segment is a uint8 with the lower 4 bits representing SegmentType and the
// upper 4 bits representing TagType.
func MakeSegment(segType SegmentType, tagType TagType) Segment {
	// Detailing what this expression does for those who do not use bitmapping often
	// enough to intuitively understand how this works:
	//
	// 1. uint8(tagType) << 4 — Shifts the TagType value left by 4 bits, positioning
	//                          it in the high 4 bits of the byte (bits 4-7).
	// 2. | uint8(segType)    — Combines the shifted TagType with the SegmentType using
	//                          a bitwise OR operation. The SegmentType occupies the
	//                          lower 4 bits (bits 0-3).
	// 3. Segment(...)        — Converts the resulting combined byte back to the
	//                          Segment type.
	//
	return Segment((uint8(tagType) << 4) | uint8(segType))
}

// MakeWhitespaceSegment creates a Segment representing whitespace content.
// It uses SegmentType=Whitespace and TagType=NotApplicable.
func MakeWhitespaceSegment() Segment {
	return MakeSegment(Whitespace, NotApplicable)
}

// MakeTextContentSegment creates a Segment representing plain text content.
// It uses SegmentType=TextContent and TagType=NotApplicable.
func MakeTextContentSegment() Segment {
	return MakeSegment(TextContent, NotApplicable)
}

// MakeTagSegment creates a Segment representing a complete tag with the
// specified tag type. It uses SegmentType=CompleteTag and the TagType provided
// as parameter.
func MakeTagSegment(tagType TagType) Segment {
	return MakeSegment(CompleteTag, tagType)
}

// MakeCommentTagSegment creates a Segment representing a comment tag with the
// specified segment type. It uses the SegmentType provided as parameter and
// TagType=CommentTag.
func MakeCommentTagSegment(segmentType SegmentType) Segment {
	return MakeSegment(segmentType, CommentTag)
}

// MakeBeginTagSegment creates a Segment representing the beginning of an
// enclosing tag. It uses SegmentType=BeginTag and the TagType provided as
// parameter.
func MakeBeginTagSegment(tagType TagType) Segment {
	return MakeSegment(BeginTag, tagType)
}

// MakeEndTagSegment creates a Segment representing the end of an enclosing tag.
// It uses SegmentType=EndTag and the TagType provided as parameter.
func MakeEndTagSegment(tagType TagType) Segment {
	return MakeSegment(EndTag, tagType)
}

// WithType creates a new Segment with the same TagType but a different
// SegmentType. This keeps the tag semantics but changes the structural role of
// the segment.
func (s Segment) WithType(typ SegmentType) Segment {
	// Detailing what this expression does for those who do not use bitmapping often
	// enough to intuitively understand how this works:
	//
	// 1. uint8(s) & 0xF0 — Preserves the high 4 bits (the TagType) of the original
	//                      segment by masking with 0xF0 (binary 11110000).
	// 2. | uint8(typ)    — Combines the preserved high bits with the new SegmentType
	//                      value in the low bits using a bitwise OR operation.
	// 3. Segment(...)    — Converts the result back to the Segment type
	//
	return Segment((uint8(s) & 0xF0) | uint8(typ))
}

// Type extracts the low 4 bits of a Segment to return the SegmentType. This
// represents the structural role of the segment (e.g., CompleteTag, BeginTag).
func (s Segment) Type() SegmentType {
	return SegmentType(s & 0x0F)
}

// TagType extracts the high 4 bits of a Segment to return the TagType. This
// represents the semantic meaning of the tag (e.g., SectionTag, VarTag).
func (s Segment) TagType() TagType {
	return TagType(s >> 4)
}

// IsValidTagType checks if the tag type embedded in the segment is valid. It
// delegates to the TagType.IsValid() method.
func (s Segment) IsValidTagType() bool {
	return s.TagType().IsValid()
}

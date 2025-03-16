package classifier

import (
	"fmt"
	"strings"
)

const NumPreallocatedSegments = 3

type Segments []Segment

func NewSegments() Segments {
	return make(Segments, 0, NumPreallocatedSegments)
}

type Segment uint8 // 4 bits for SegmentType, 4 bits for TagType

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

func (ss *Segments) Reset() {
	if len(*ss) == 0 {
		goto end
	}
	(*ss)[0] = 0 // Reset to zero value
end:
	return
}
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

// MakeSegment creates a bitmapped value combining SegmentType and TagType
// A segment is a uint8 with the lower 4 bits representing SegmentType and the upper 4 bits representing TagType
// Helper functions to pack/unpack Segment
func MakeSegment(segType SegmentType, tagType TagType) Segment {
	// 1. uint8(tagType) << 4 — Shifts the TagType value left by 4 bits, positioning it in the high 4 bits of the byte (bits 4-7).
	// 2. | uint8(segType)    — Combines the shifted TagType with the SegmentType using a bitwise OR operation. The SegmentType occupies the low 4 bits (bits 0-3).
	// 3. Segment(...)        — Converts the resulting combined byte back to the Segment type.
	return Segment((uint8(tagType) << 4) | uint8(segType))
}

// MakeWhitespaceSegment is a convenience functions to create a Segment for
// SegmentType=Whitespace and TagType=NotApplicable
func MakeWhitespaceSegment() Segment {
	return MakeSegment(Whitespace, NotApplicable)
}

// MakeTextContentSegment is a convenience functions to create a Segment for
// SegmentType=TextContent and TagType=NotApplicable
func MakeTextContentSegment() Segment {
	return MakeSegment(TextContent, NotApplicable)
}

// MakeTagSegment is a convenience functions to create a Segment for
// SegmentType=CompleteTag and TagType=<tagType> parameter
func MakeTagSegment(tagType TagType) Segment {
	return MakeSegment(CompleteTag, tagType)
}

// MakeCommentTagSegment is a convenience functions to create a Segment for
// SegmentType=<segmentType> and TagType=CommentTag
func MakeCommentTagSegment(segmentType SegmentType) Segment {
	return MakeSegment(segmentType, CommentTag)
}

// MakeBeginTagSegment is a convenience functions to create a Segment for
// SegmentType=BeginTag and TagType=<tagType> parameter
func MakeBeginTagSegment(tagType TagType) Segment {
	return MakeSegment(BeginTag, tagType)
}

// MakeEndTagSegment is a convenience functions to create a Segment for
// SegmentType=EndTag and TagType=<tagType> parameter
func MakeEndTagSegment(tagType TagType) Segment {
	return MakeSegment(EndTag, tagType)
}

// WithType creates a new Segment with the same TagType but a different SegmentType
func (s Segment) WithType(typ SegmentType) Segment {
	// 1. uint8(s) & 0xF0 — Preserves the high 4 bits (the TagType) of the original segment by masking with 0xF0 (binary 11110000).
	// 2. | uint8(typ)    — Combines the preserved high bits with the new SegmentType value in the low bits using a bitwise OR operation.
	// 3. Segment(...)    — Converts the result back to the Segment type
	return Segment((uint8(s) & 0xF0) | uint8(typ))
}

// Type extracts the low 4 bits of a Segment to return a SegmentType.
func (s Segment) Type() SegmentType {
	return SegmentType(s & 0x0F)
}

// TagType extracts the high 4 bits of a Segment to return a TagType.
func (s Segment) TagType() TagType {
	return TagType(s >> 4)
}

// IsValidTagType checks if the tag type embedded in the segment is valid
func (s Segment) IsValidTagType() bool {
	return s.TagType().IsValid()
}

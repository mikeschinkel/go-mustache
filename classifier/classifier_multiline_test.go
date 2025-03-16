package classifier

import (
	"testing"
)

func TestClassifierWithTestCaseSet3(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Multiple lines with one standalone tag",
			template: "Text\n{{tag}}\nMore text",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Only the middle line should be identified as standalone",
		},
		{
			name:     "Multiple lines with indented standalone tag",
			template: "Text\n    {{tag}}\nMore text",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				2: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Indented tag on its own line is standalone",
		},
		{
			name:     "Multiple lines with multiple standalone tags",
			template: "{{tag1}}\nText\n{{tag2}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "First and last lines should be identified as standalone",
		},
		{
			name:     "Multiple lines with multiple standalone tags and trailing newline",
			template: "{{tag1}}\nText\n{{tag2}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: StandaloneLine, Segments: varTagSegments},
				3: {Type: EmptyLine},
			},
			comment: "First and third lines should be identified as standalone; trailing newline creates an empty line",
		},
		{
			name:     "Empty lines between tags",
			template: "{{tag1}}\n\n{{tag2}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: EmptyLine},
				2: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Empty lines should be marked as empty lines",
		},
		{
			name:     "Empty lines between tags with trailing newline",
			template: "{{tag1}}\n\n{{tag2}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: EmptyLine},
				2: {Type: StandaloneLine, Segments: varTagSegments},
				3: {Type: EmptyLine},
			},
			comment: "Empty lines should be marked as empty lines; trailing newline creates an empty line",
		},
	})
}

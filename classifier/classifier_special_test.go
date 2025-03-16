package classifier

import (
	"strings"
	"testing"
)

func TestClassifierWithSpecialScenarios(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Empty template",
			template: "",
			want: Lines{
				0: {Type: EmptyLine},
			},
			comment: "Empty templates should return an empty slice",
		},
		{
			name:     "Single line, no tag",
			template: "Just some text",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Lines without tags are non-tag lines",
		},
		{
			name:     "Single line, no tag, with newline",
			template: "Just some text\n",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: EmptyLine},
			},
			comment: "Lines without tags are non-tag lines; trailing newline creates an empty line",
		},
		{
			name:     "Line with only whitespace",
			template: "    \n",
			want: Lines{
				0: {Type: WhitespaceLine, Segments: whitespaceSegments},
				1: {Type: EmptyLine},
			},
			comment: "Lines with only whitespace are whitespace lines; trailing newline creates an empty line",
		},
		{
			name:     "Tag at very beginning of long template",
			template: "{{tag}}\n" + strings.Repeat("x ", 1000) + "\n{{tag}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Large templates should still correctly identify standalone tags",
		},
		{
			name:     "Tag at very beginning of long template with trailing newline",
			template: "{{tag}}\n" + strings.Repeat("x ", 1000) + "\n{{tag}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: StandaloneLine, Segments: varTagSegments},
				3: {Type: EmptyLine},
			},
			comment: "Large templates should still correctly identify standalone tags; trailing newline creates an empty line",
		},
		{
			name:     "Tags after ASCII art",
			template: "  /\\\n /  \\\n{{tag}}\n",
			want: Lines{
				0: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: varTagSegments},
				3: {Type: EmptyLine},
			},
			comment: "ASCII art lines are non-tag lines, but the tag line is standalone; trailing newline creates an empty line",
		},
	})
}

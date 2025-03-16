package test

import (
	"strings"
	"testing"
)

func TestClassifierWithStandaloneTags(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Simple standalone tag",
			template: "{{tag}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Tag alone on a line is standalone",
		},
		{
			name:     "Simple standalone tag with newline",
			template: "{{tag}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: EmptyLine},
			},
			comment: "Tag alone on a line followed by newline creates a standalone tag line, plus an empty line",
		},
		{
			name:     "Indented standalone tag",
			template: "    {{tag}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
			},
			comment: "Tag with only whitespace before it is standalone",
		},
		{
			name:     "Indented standalone tag with newline",
			template: "    {{tag}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Tag with only whitespace before it is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Standalone tag with trailing whitespace",
			template: "{{tag}}    ",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{varTagSegment, whitespaceSegment}},
			},
			comment: "Tag with only whitespace after it is standalone",
		},
		{
			name:     "Standalone tag with trailing whitespace and newline",
			template: "{{tag}}    \n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Tag with only whitespace after it is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Very long line with only tag",
			template: strings.Repeat(" ", 100) + "{{tag}}" + strings.Repeat(" ", 100),
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
			},
			comment: "Long line with only whitespace and a tag is standalone",
		},
		{
			name:     "Very long line with only tag and trailing newline",
			template: strings.Repeat(" ", 100) + "{{tag}}" + strings.Repeat(" ", 100) + "\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Long line with only whitespace and a tag is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Indented standalone tag with trailing whitespace",
			template: "    {{tag}}    ",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
			},
			comment: "Tag with whitespace before and after is standalone",
		},
		{
			name:     "Indented standalone tag with trailing whitespace and newline",
			template: "    {{tag}}    \n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Tag with whitespace before and after is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag followed by newline and whitespace",
			template: "{{tag}}\n    ",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: WhitespaceLine, Segments: whitespaceSegments},
			},
			comment: "First line is standalone, line with only whitespace is a whitespace line",
		},
		{
			name:     "Tag surrounded by tabs",
			template: "\t\t\t{{tag}}\t\t\t",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
			},
			comment: "Tabs count as whitespace for standalone detection",
		},
		{
			name:     "Tag surrounded by tabs with trailing newline",
			template: "\t\t\t{{tag}}\t\t\t\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Tabs count as whitespace for standalone detection; trailing newline creates an empty line",
		},
		{
			name:     "Mixed tabs and spaces",
			template: " \t \t{{tag}}\t \t ",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
			},
			comment: "Mix of tabs and spaces counts as whitespace",
		},
		{
			name:     "Mixed tabs and spaces with trailing newline",
			template: " \t \t{{tag}}\t \t \n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Mix of tabs and spaces counts as whitespace; trailing newline creates an empty line",
		},
		{
			name:     "Unescaped variable tags",
			template: "{{{unescaped}}}\n{{&alsoUnescaped}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: tripleBraceUnescapedSegments},
				1: {Type: StandaloneLine, Segments: ampersandUnescapedSegments},
			},
			comment: "Unescaped variable tags can be standalone",
		},
		{
			name:     "Unescaped variable tags with trailing newline",
			template: "{{{unescaped}}}\n{{&alsoUnescaped}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: tripleBraceUnescapedSegments},
				1: {Type: StandaloneLine, Segments: ampersandUnescapedSegments},
				2: {Type: EmptyLine},
			},
			comment: "Unescaped variable tags can be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Template with unusual spacing",
			template: "{{ tag }}\n{{  tag  }}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Internal whitespace in tags doesn't affect standalone status",
		},
		{
			name:     "Template with unusual spacing and trailing newline",
			template: "{{ tag }}\n{{  tag  }}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: EmptyLine},
			},
			comment: "Internal whitespace in tags doesn't affect standalone status; trailing newline creates an empty line",
		},
		{
			name:     "Template with unusual newlines",
			template: "{{tag}}\r\n{{tag}}\r{{tag}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Different newline styles (LF, CRLF, CR) should be handled",
		},
		{
			name:     "Inverted Section - No content",
			template: "{{^section}}\n{{/section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
			},
			comment: "Inverted section without content should be standalone",
		},
		{
			name:     "Inverted Section - No content with trailing newline",
			template: "{{^section}}\n{{/section}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
				2: {Type: EmptyLine},
			},
			comment: "Inverted section without content should be standalone with trailing newline",
		},
	})
}

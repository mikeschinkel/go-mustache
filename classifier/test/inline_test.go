package test

import (
	"testing"
)

func TestClassifierWithInlineTags(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Single line with tag",
			template: "Text {{tag}} more text",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
			},
			comment: "Tags surrounded by text are inline tags",
		},
		{
			name:     "Single line with tag and newline",
			template: "Text {{tag}} more text\n",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Tags surrounded by text are inline tags; trailing newline creates an empty line",
		},
		{
			name:     "Tag at beginning of line",
			template: "{{tag}} text",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{varTagSegment, textContentSegment}},
			},
			comment: "Tags with text after them are inline tags",
		},
		{
			name:     "Tag at end of line",
			template: "text {{tag}}",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment}},
			},
			comment: "Tags with text before them are inline tags",
		},
		{
			name:     "Very long line with tag at end",
			template: "This is a very long line with lots of text that continues for a while and then eventually has a tag at the end {{tag}}",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment}},
			},
			comment: "Long line with tag at end is an inline tag",
		},
		{
			name:     "Very long line with tag at end and trailing newline",
			template: "This is a very long line with lots of text that continues for a while and then eventually has a tag at the end {{tag}}\n",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Long line with tag at end is an inline tag; trailing newline creates an empty line",
		},
		{
			name:     "Tag with whitespace at start but text after",
			template: "    {{tag}} text",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{whitespaceSegment, varTagSegment, textContentSegment}},
			},
			comment: "Inline tag due to text after the tag",
		},
		{
			name:     "Tag with whitespace at start but text after and newline",
			template: "    {{tag}} text\n",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{whitespaceSegment, varTagSegment, textContentSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Inline tag due to text after the tag; trailing newline creates an empty line",
		},
		{
			name:     "Template with no newlines",
			template: "{{#section}}{{>partial}}{{/section}}",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{sectionBeginSegment, partialTagSegment, sectionEndSegment}},
			},
			comment: "One line template with multiple tags is an inline tag line",
		},
		{
			name:     "Tag with Unicode whitespace",
			template: "\u2003\u2002{{tag}}\u2003\u2002",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
			}, // Most implementations only treat ASCII whitespace as whitespace
			comment: "Unicode whitespace characters might not be treated as whitespace in some implementations",
		},
		{
			name:     "Tag with Unicode whitespace and trailing newline",
			template: "\u2003\u2002{{tag}}\u2003\u2002\n",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Unicode whitespace characters might not be treated as whitespace; trailing newline creates an empty line",
		},
		{
			name:     "Multiple tags on one line",
			template: "{{tag1}}{{tag2}}",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{varTagSegment, varTagSegment}},
			},
			comment: "Multiple tags on one line means it's an inline tag line",
		},
		{
			name:     "Multiple tags with whitespace between",
			template: "{{tag1}}    {{tag2}}",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{varTagSegment, whitespaceSegment, varTagSegment}},
			},
			comment: "Multiple tags on one line with whitespace between means it's an inline tag line",
		},
	})
}

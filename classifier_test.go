package mustache

import (
	"testing"
)

func TestDetectLineTypes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     LineNoTypes
		comment  string
	}{
		{
			name:     "Empty template",
			template: "",
			want:     LineNoTypes{{EmptyLine}},
			comment:  "Empty templates should return an empty slice",
		},
		{
			name:     "Single line, no tag",
			template: "Just some text",
			want:     LineNoTypes{{TextLine}},
			comment:  "Lines without tags are non-tag lines",
		},
		{
			name:     "Single line, no tag, with newline",
			template: "Just some text\n",
			want:     LineNoTypes{{TextLine}, {EmptyLine}},
			comment:  "Lines without tags are non-tag lines; trailing newline creates an empty line",
		},
		{
			name:     "Single line with tag",
			template: "Text {{tag}} more text",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Tags surrounded by text are inline tags",
		},
		{
			name:     "Single line with tag and newline",
			template: "Text {{tag}} more text\n",
			want:     LineNoTypes{{InlineTagLine}, {EmptyLine}},
			comment:  "Tags surrounded by text are inline tags; trailing newline creates an empty line",
		},
		{
			name:     "Tag at beginning of line",
			template: "{{tag}} text",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Tags with text after them are inline tags",
		},
		{
			name:     "Tag at end of line",
			template: "text {{tag}}",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Tags with text before them are inline tags",
		},
		{
			name:     "Simple standalone tag",
			template: "{{tag}}",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Tag alone on a line is standalone",
		},
		{
			name:     "Simple standalone tag with newline",
			template: "{{tag}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Tag alone on a line followed by newline creates a standalone tag line, plus an empty line",
		},
		{
			name:     "Indented standalone tag",
			template: "    {{tag}}",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Tag with only whitespace before it is standalone",
		},
		{
			name:     "Indented standalone tag with newline",
			template: "    {{tag}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Tag with only whitespace before it is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Standalone tag with trailing whitespace",
			template: "{{tag}}    ",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Tag with only whitespace after it is standalone",
		},
		{
			name:     "Standalone tag with trailing whitespace and newline",
			template: "{{tag}}    \n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Tag with only whitespace after it is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Indented standalone tag with trailing whitespace",
			template: "    {{tag}}    ",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Tag with whitespace before and after is standalone",
		},
		{
			name:     "Indented standalone tag with trailing whitespace and newline",
			template: "    {{tag}}    \n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Tag with whitespace before and after is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Multiple lines with one standalone tag",
			template: "Text\n{{tag}}\nMore text",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {TextLine}},
			comment:  "Only the middle line should be identified as standalone",
		},
		{
			name:     "Multiple lines with indented standalone tag",
			template: "Text\n    {{tag}}\nMore text",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {TextLine}},
			comment:  "Indented tag on its own line is standalone",
		},
		{
			name:     "Multiple lines with multiple standalone tags",
			template: "{{tag1}}\nText\n{{tag2}}",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {StandaloneTagLine}},
			comment:  "First and last lines should be identified as standalone",
		},
		{
			name:     "Multiple lines with multiple standalone tags and trailing newline",
			template: "{{tag1}}\nText\n{{tag2}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "First and third lines should be identified as standalone; trailing newline creates an empty line",
		},
		{
			name:     "Multiple tags on one line",
			template: "{{tag1}}{{tag2}}",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Multiple tags on one line means it's an inline tag line",
		},
		{
			name:     "Multiple tags with whitespace between",
			template: "{{tag1}}    {{tag2}}",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Multiple tags on one line with whitespace between means it's an inline tag line",
		},
		{
			name:     "Tag followed by newline and whitespace",
			template: "{{tag}}\n    ",
			want:     LineNoTypes{{StandaloneTagLine}, {WhitespaceLine}},
			comment:  "First line is standalone, line with only whitespace is a whitespace line",
		},
		{
			name:     "Empty lines between tags",
			template: "{{tag1}}\n\n{{tag2}}",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}, {StandaloneTagLine}},
			comment:  "Empty lines should be marked as empty lines",
		},
		{
			name:     "Empty lines between tags with trailing newline",
			template: "{{tag1}}\n\n{{tag2}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Empty lines should be marked as empty lines; trailing newline creates an empty line",
		},
		{
			name:     "Line with only whitespace",
			template: "    \n",
			want:     LineNoTypes{{WhitespaceLine}, {EmptyLine}},
			comment:  "Lines with only whitespace are whitespace lines; trailing newline creates an empty line",
		},
		{
			name:     "Tag with whitespace at start but text after",
			template: "    {{tag}} text",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Inline tag due to text after the tag",
		},
		{
			name:     "Tag with whitespace at start but text after and newline",
			template: "    {{tag}} text\n",
			want:     LineNoTypes{{InlineTagLine}, {EmptyLine}},
			comment:  "Inline tag due to text after the tag; trailing newline creates an empty line",
		},
		{
			name:     "Complex multi-line template",
			template: "Before\n{{#section}}\n  {{name}}\n{{/section}}\nAfter",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}},
			comment:  "Section opening and closing tags should be standalone",
		},
		{
			name:     "Complex multi-line template with trailing newline",
			template: "Before\n{{#section}}\n  {{name}}\n{{/section}}\nAfter\n",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}, {EmptyLine}},
			comment:  "Section opening and closing tags should be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Different tag types in standalone positions",
			template: "{{#section}}\n{{>partial}}\n{{/section}}",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Different tag types can all be standalone",
		},
		{
			name:     "Different tag types in standalone positions with trailing newline",
			template: "{{#section}}\n{{>partial}}\n{{/section}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Different tag types can all be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag inside HTML",
			template: "<div>\n  {{tag}}\n</div>",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {TextLine}},
			comment:  "Tag on its own line inside HTML is standalone",
		},
		{
			name:     "Tag inside HTML with trailing newline",
			template: "<div>\n  {{tag}}\n</div>\n",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {TextLine}, {EmptyLine}},
			comment:  "Tag on its own line inside HTML is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Line ends with standalone tag but no newline",
			template: "Text\n{{tag}}",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}},
			comment:  "Last line containing only a tag is standalone even without trailing newline",
		},
		{
			name:     "Tag preceded by newline and followed by newline",
			template: "\n{{tag}}\n",
			want:     LineNoTypes{{EmptyLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Empty lines before and after a standalone tag",
		},
		{
			name:     "Tag with leading and trailing newlines",
			template: "\n\n    {{tag}}    \n\n",
			want:     LineNoTypes{{EmptyLine}, {EmptyLine}, {StandaloneTagLine}, {EmptyLine}, {EmptyLine}},
			comment:  "Tag with whitespace before and after, surrounded by empty lines",
		},
		{
			name:     "Real-world parent/block pattern",
			template: "{{<parent}}\n{{$block}}Content{{/block}}\n{{/parent}}",
			want:     LineNoTypes{{StandaloneTagLine}, {InlineTagLine}, {StandaloneTagLine}},
			comment:  "Parent tag and closing tag are standalone, block content line is inline",
		},
		{
			name:     "Real-world parent/block pattern with trailing newline",
			template: "{{<parent}}\n{{$block}}Content{{/block}}\n{{/parent}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {InlineTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Parent tag and closing tag are standalone, block content line is inline; trailing newline creates an empty line",
		},
		{
			name:     "Real-world indented template",
			template: "<html>\n  <body>\n    {{>header}}\n    <main>\n      {{#items}}\n        {{.}}\n      {{/items}}\n    </main>\n  </body>\n</html>",
			want:     LineNoTypes{{TextLine}, {TextLine}, {StandaloneTagLine}, {TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}, {TextLine}, {TextLine}},
			comment:  "Partial and section tags should be detected as standalone",
		},
		{
			name:     "Real-world indented template with trailing newline",
			template: "<html>\n  <body>\n    {{>header}}\n    <main>\n      {{#items}}\n        {{.}}\n      {{/items}}\n    </main>\n  </body>\n</html>\n",
			want:     LineNoTypes{{TextLine}, {TextLine}, {StandaloneTagLine}, {TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}, {TextLine}, {TextLine}, {EmptyLine}},
			comment:  "Partial and section tags should be detected as standalone; trailing newline creates an empty line",
		},
		{
			name:     "Partial with complex indentation",
			template: "<div>\n    {{>partial}}\n        {{>partial}}\n</div>",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}},
			comment:  "Partial tags with different indentation levels",
		},
		{
			name:     "Partial with complex indentation with trailing newline",
			template: "<div>\n    {{>partial}}\n        {{>partial}}\n</div>\n",
			want:     LineNoTypes{{TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {TextLine}, {EmptyLine}},
			comment:  "Partial tags with different indentation levels; trailing newline creates an empty line",
		},
		{
			name:     "Template with no newlines",
			template: "{{#section}}{{>partial}}{{/section}}",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "One line template with multiple tags is an inline tag line",
		},
		{
			name:     "Template with unusual spacing",
			template: "{{ tag }}\n{{  tag  }}",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Internal whitespace in tags doesn't affect standalone status",
		},
		{
			name:     "Template with unusual spacing and trailing newline",
			template: "{{ tag }}\n{{  tag  }}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Internal whitespace in tags doesn't affect standalone status; trailing newline creates an empty line",
		},
		{
			name:     "Tags after ASCII art",
			template: "  /\\\n /  \\\n{{tag}}\n",
			want:     LineNoTypes{{TextLine}, {TextLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "ASCII art lines are non-tag lines, but the tag line is standalone; trailing newline creates an empty line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewTemplateClassifier(tt.template)
			got, err := classifier.Classify()
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !got.Equal(tt.want) {
				showMismatch(t, got, tt.want, tt.template, tt.comment)
			}
		})
	}
}

// TestComplexTemplates contains additional tests with more complex templates
func TestComplexTemplates(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     LineNoTypes
		comment  string
	}{
		{
			name:     "Multi-block template inheritance pattern",
			template: "{{<layout}}\n{{$title}}Page Title{{/title}}\n{{$content}}\n  <p>Page content</p>\n  <ul>\n    <li>Item 1</li>\n    <li>Item 2</li>\n  </ul>\n{{/content}}\n{{/layout}}",
			want:     LineNoTypes{{StandaloneTagLine}, {InlineTagLine}, {StandaloneTagLine}, {TextLine}, {TextLine}, {TextLine}, {TextLine}, {TextLine}, {StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Complex template inheritance with nested content blocks",
		},
		{
			name:     "Multi-block template inheritance pattern with trailing newline",
			template: "{{<layout}}\n{{$title}}Page Title{{/title}}\n{{$content}}\n  <p>Page content</p>\n  <ul>\n    <li>Item 1</li>\n    <li>Item 2</li>\n  </ul>\n{{/content}}\n{{/layout}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {InlineTagLine}, {StandaloneTagLine}, {TextLine}, {TextLine}, {TextLine}, {TextLine}, {TextLine}, {StandaloneTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Complex template inheritance with nested content blocks; trailing newline creates an empty line",
		},
		{
			name:     "Nested sections with indentation",
			template: "{{#outer}}\n  {{#inner}}\n    {{value}}\n  {{/inner}}\n{{/outer}}",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Nested sections with consistent indentation should all be standalone",
		},
		{
			name:     "Nested sections with indentation and trailing newline",
			template: "{{#outer}}\n  {{#inner}}\n    {{value}}\n  {{/inner}}\n{{/outer}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Nested sections with consistent indentation should all be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag with dot notation and special characters",
			template: "{{user.name}}\n{{#user.settings.notifications}}\n  {{user.email}}\n{{/user.settings.notifications}}",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Tags with complex paths and dot notation can be standalone",
		},
		{
			name:     "Tag with dot notation and special characters with trailing newline",
			template: "{{user.name}}\n{{#user.settings.notifications}}\n  {{user.email}}\n{{/user.settings.notifications}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Tags with complex paths and dot notation can be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Comment tags",
			template: "{{! This is a comment }}\nContent\n{{! \n  Multi-line\n  comment\n}}\nMore content",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {MultilineTagBegin, StandaloneTagLine}, {MultilineTagMiddle}, {MultilineTagMiddle}, {MultilineTagEnd, StandaloneTagLine}, {TextLine}},
			comment:  "Comment tags, including multi-line comments, with proper multi-line tag designation",
		},
		{
			name:     "Comment tags with trailing newline",
			template: "{{! This is a comment }}\nContent\n{{! \n  Multi-line\n  comment\n}}\nMore content\n",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {MultilineTagBegin, StandaloneTagLine}, {MultilineTagMiddle}, {MultilineTagMiddle}, {MultilineTagEnd, StandaloneTagLine}, {TextLine}, {EmptyLine}},
			comment:  "Comment tags, including multi-line comments, with proper multi-line tag designation; trailing newline creates an empty line",
		},
		{
			name:     "Unescaped variable tags",
			template: "{{{unescaped}}}\n{{&alsoUnescaped}}",
			want:     LineNoTypes{{StandaloneTagLine, TripleBraceUnescaped}, {AmpersandUnescaped, StandaloneTagLine}},
			comment:  "Unescaped variable tags can be standalone",
		},
		{
			name:     "Unescaped variable tags with trailing newline",
			template: "{{{unescaped}}}\n{{&alsoUnescaped}}\n",
			want:     LineNoTypes{{StandaloneTagLine, TripleBraceUnescaped}, {AmpersandUnescaped, StandaloneTagLine}, {EmptyLine}},
			comment:  "Unescaped variable tags can be standalone; trailing newline creates an empty line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewTemplateClassifier(tt.template)
			got, err := classifier.Classify()
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !got.Equal(tt.want) {
				showMismatch(t, got, tt.want, tt.template, tt.comment)
			}
		})
	}
}

// TestEdgeCases contains tests for unusual or edge-case scenarios
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     LineNoTypes
		comment  string
	}{
		{
			name:     "Very long line with tag at end",
			template: "This is a very long line with lots of text that continues for a while and then eventually has a tag at the end {{tag}}",
			want:     LineNoTypes{{InlineTagLine}},
			comment:  "Long line with tag at end is an inline tag",
		},
		{
			name:     "Very long line with tag at end and trailing newline",
			template: "This is a very long line with lots of text that continues for a while and then eventually has a tag at the end {{tag}}\n",
			want:     LineNoTypes{{InlineTagLine}, {EmptyLine}},
			comment:  "Long line with tag at end is an inline tag; trailing newline creates an empty line",
		},
		{
			name:     "Very long line with only tag",
			template: "                                                                                {{tag}}                                                                                ",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Long line with only whitespace and a tag is standalone",
		},
		{
			name:     "Very long line with only tag and trailing newline",
			template: "                                                                                {{tag}}                                                                                \n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Long line with only whitespace and a tag is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag surrounded by tabs",
			template: "\t\t\t{{tag}}\t\t\t",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Tabs count as whitespace for standalone detection",
		},
		{
			name:     "Tag surrounded by tabs with trailing newline",
			template: "\t\t\t{{tag}}\t\t\t\n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Tabs count as whitespace for standalone detection; trailing newline creates an empty line",
		},
		{
			name:     "Mixed tabs and spaces",
			template: " \t \t{{tag}}\t \t ",
			want:     LineNoTypes{{StandaloneTagLine}},
			comment:  "Mix of tabs and spaces counts as whitespace",
		},
		{
			name:     "Mixed tabs and spaces with trailing newline",
			template: " \t \t{{tag}}\t \t \n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}},
			comment:  "Mix of tabs and spaces counts as whitespace; trailing newline creates an empty line",
		},
		{
			name:     "Tag with Unicode whitespace",
			template: "\u2003\u2002{{tag}}\u2003\u2002",
			want:     LineNoTypes{{InlineTagLine}}, // Most implementations only treat ASCII whitespace as whitespace
			comment:  "Unicode whitespace characters might not be treated as whitespace in some implementations",
		},
		{
			name:     "Tag with Unicode whitespace and trailing newline",
			template: "\u2003\u2002{{tag}}\u2003\u2002\n",
			want:     LineNoTypes{{InlineTagLine}, {EmptyLine}},
			comment:  "Unicode whitespace characters might not be treated as whitespace; trailing newline creates an empty line",
		},
		{
			name:     "Tag at very beginning of long template",
			template: "{{tag}}\n" + string(make([]byte, 10000)) + "\n{{tag}}",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {StandaloneTagLine}},
			comment:  "Large templates should still correctly identify standalone tags",
		},
		{
			name:     "Tag at very beginning of long template with trailing newline",
			template: "{{tag}}\n" + string(make([]byte, 10000)) + "\n{{tag}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {TextLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Large templates should still correctly identify standalone tags; trailing newline creates an empty line",
		},
		{
			name:     "Template with unusual newlines",
			template: "{{tag}}\r\n{{tag}}\r{{tag}}",
			want:     LineNoTypes{{StandaloneTagLine}, {StandaloneTagLine}, {StandaloneTagLine}},
			comment:  "Different newline styles (LF, CRLF, CR) should be handled",
		},
		{
			name:     "Line with just closing delimiter",
			template: "{{\ntag\n}}",
			want:     LineNoTypes{{MultilineTagBegin, StandaloneTagLine}, {MultilineTagMiddle}, {MultilineTagEnd, StandaloneTagLine}},
			comment:  "Multi-line tag contents are properly identified as beginning, middle, and end",
		},
		{
			name:     "Multiple sequential newlines",
			template: "{{tag}}\n\n\n{{tag}}",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}, {EmptyLine}, {StandaloneTagLine}},
			comment:  "Multiple empty lines between tags",
		},
		{
			name:     "Multiple sequential newlines with trailing newline",
			template: "{{tag}}\n\n\n{{tag}}\n",
			want:     LineNoTypes{{StandaloneTagLine}, {EmptyLine}, {EmptyLine}, {StandaloneTagLine}, {EmptyLine}},
			comment:  "Multiple empty lines between tags; trailing newline creates an empty line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewTemplateClassifier(tt.template)
			got, err := classifier.Classify()
			if err != nil {
				t.Error(err.Error())
				return
			}
			if !got.Equal(tt.want) {
				showMismatch(t, got, tt.want, tt.template, tt.comment)
			}
		})
	}
}

func showMismatch(t *testing.T, got, want any, template, comment string) {
	t.Errorf("ERROR:"+
		"\n\tGot:      %v"+
		"\n\tWant:     %v"+
		"\n\tTemplate: %q"+
		"\n\tComment:  %s\n",
		got,
		want,
		template,
		comment,
	)
}

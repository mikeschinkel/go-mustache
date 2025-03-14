package classifier

import (
	"strings"
	"testing"
)

var (
	whitespaceSegment           = Segment{Type: Whitespace}
	textContentSegment          = Segment{Type: TextContent}
	varTagSegment               = Segment{TagType: VarTag}
	sectionBeginSegment         = Segment{TagType: SectionTag, Type: BeginTag}
	sectionEndSegment           = Segment{TagType: SectionTag, Type: EndTag}
	sectionBeginInvertedSegment = Segment{TagType: InvertedSectionTag, Type: BeginTag}
	sectionEndInvertedSegment   = Segment{TagType: InvertedSectionTag, Type: EndTag}
	partialTagSegment           = Segment{TagType: PartialTag}
	parentBeginSegment          = Segment{TagType: ParentTag, Type: BeginTag}
	parentEndSegment            = Segment{TagType: ParentTag, Type: EndTag}
	blockBeginSegment           = Segment{TagType: BlockTag, Type: BeginTag}
	blockEndSegment             = Segment{TagType: BlockTag, Type: EndTag}
	delimiterTagSegment         = Segment{TagType: DelimiterTag}
	commentTagSegment           = Segment{TagType: CommentTag}

	tripleBraceUnescapedSegment = Segment{TagType: TripleBraceUnescaped}
	ampersandUnescapedSegment   = Segment{TagType: AmpersandUnescaped}

	tripleBraceUnescapedSegments = Segments{tripleBraceUnescapedSegment}
	ampersandUnescapedSegments   = Segments{ampersandUnescapedSegment}

	whitespaceSegments           = Segments{whitespaceSegment}
	varTagSegments               = Segments{varTagSegment}
	textContentSegments          = Segments{textContentSegment}
	delimiterTagSegments         = Segments{delimiterTagSegment}
	sectionBeginSegments         = Segments{sectionBeginSegment}
	sectionEndSegments           = Segments{sectionEndSegment}
	sectionBeginInvertedSegments = Segments{sectionBeginInvertedSegment}
	sectionEndInvertedSegments   = Segments{sectionEndInvertedSegment}
	partialTagSegments           = Segments{partialTagSegment}
	parentBeginSegments          = Segments{parentBeginSegment}
	parentEndSegments            = Segments{parentEndSegment}
	blockBeginSegments           = Segments{blockBeginSegment}
	blockEndSegments             = Segments{blockEndSegment}
	commentTagSegments           = Segments{commentTagSegment}

	multilineCommentBegin   = Segment{Type: MultilineBegin, TagType: CommentTag}
	multilineCommentMiddle  = Segment{Type: MultilineMiddle, TagType: CommentTag}
	multilineCommentEnd     = Segment{Type: MultilineEnd, TagType: CommentTag}
	multilineCommentBegins  = Segments{multilineCommentBegin}
	multilineCommentMiddles = Segments{multilineCommentMiddle}
	multilineCommentEnds    = Segments{multilineCommentEnd}
)

func TestTemplateClassifier(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     Lines
		error    string
		comment  string
	}{
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
			name:     "Complex multi-line template",
			template: "Before\n{{#section}}\n  {{name}}\n{{/section}}\nAfter",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: sectionBeginSegments},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: sectionEndSegments},
				4: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Section opening and closing tags should be standalone",
		},
		{
			name:     "Complex multi-line template with trailing newline",
			template: "Before\n{{#section}}\n  {{name}}\n{{/section}}\nAfter\n",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: sectionBeginSegments},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: sectionEndSegments},
				4: {Type: TextLine, Segments: textContentSegments},
				5: {Type: EmptyLine},
			},
			comment: "Section opening and closing tags should be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Different tag types in standalone positions",
			template: "{{#section}}\n{{>partial}}\n{{/section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: partialTagSegments},
				2: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			comment: "Different tag types can all be standalone",
		},
		{
			name:     "Different tag types in standalone positions with trailing newline",
			template: "{{#section}}\n{{>partial}}\n{{/section}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: partialTagSegments},
				2: {Type: StandaloneLine, Segments: sectionEndSegments},
				3: {Type: EmptyLine},
			},
			comment: "Different tag types can all be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag inside HTML",
			template: "<div>\n  {{tag}}\n</div>",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				2: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Tag on its own line inside HTML is standalone",
		},
		{
			name:     "Tag inside HTML with trailing newline",
			template: "<div>\n  {{tag}}\n</div>\n",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				2: {Type: TextLine, Segments: textContentSegments},
				3: {Type: EmptyLine},
			},
			comment: "Tag on its own line inside HTML is standalone; trailing newline creates an empty line",
		},
		{
			name:     "Line ends with standalone tag but no newline",
			template: "Text\n{{tag}}",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Last line containing only a tag is standalone even without trailing newline",
		},
		{
			name:     "Tag preceded by newline and followed by newline",
			template: "\n{{tag}}\n",
			want: Lines{
				0: {Type: EmptyLine},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: EmptyLine}},
			comment: "Empty lines before and after a standalone tag",
		},
		{
			name:     "Tag with leading and trailing newlines",
			template: "\n\n    {{tag}}    \n\n",
			want: Lines{
				0: {Type: EmptyLine},
				1: {Type: EmptyLine},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				3: {Type: EmptyLine},
				4: {Type: EmptyLine},
			},
			comment: "Tag with whitespace before and after, surrounded by empty lines",
		},
		{
			name:     "Real-world parent/block pattern",
			template: "{{<parent}}\n{{$block}}Content{{/block}}\n{{/parent}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: parentBeginSegments},
				1: {Type: InlineLine, Segments: Segments{blockBeginSegment, textContentSegment, blockEndSegment}},
				2: {Type: StandaloneLine, Segments: parentEndSegments},
			},
			comment: "Parent tag and closing tag are standalone, block content line is inline",
		},
		{
			name:     "Real-world parent/block pattern with trailing newline",
			template: "{{<parent}}\n{{$block}}Content{{/block}}\n{{/parent}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: parentBeginSegments},
				1: {Type: InlineLine, Segments: Segments{blockBeginSegment, textContentSegment, blockEndSegment}},
				2: {Type: StandaloneLine, Segments: parentEndSegments},
				3: {Type: EmptyLine}},
			comment: "Parent tag and closing tag are standalone, block content line is inline; trailing newline creates an empty line",
		},
		{
			name:     "Real-world indented template",
			template: "<html>\n  <body>\n    {{>header}}\n    <main>\n      {{#items}}\n        {{.}}\n      {{/items}}\n    </main>\n  </body>\n</html>",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				3: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				4: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: SectionTag, Type: BeginTag}}},
				5: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: DotTag}}},
				6: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: SectionTag, Type: EndTag}}},
				7: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				8: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				9: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Partial and section tags should be detected as standalone",
		},
		{
			name:     "Real-world indented template with trailing newline",
			template: "<html>\n  <body>\n    {{>header}}\n    <main>\n      {{#items}}\n        {{.}}\n      {{/items}}\n    </main>\n  </body>\n</html>\n",
			want: Lines{
				0:  {Type: TextLine, Segments: textContentSegments},
				1:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2:  {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				3:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				4:  {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: SectionTag, Type: BeginTag}}},
				5:  {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: DotTag}}},
				6:  {Type: StandaloneLine, Segments: Segments{whitespaceSegment, {TagType: SectionTag, Type: EndTag}}},
				7:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				8:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				9:  {Type: TextLine, Segments: textContentSegments},
				10: {Type: EmptyLine},
			},
			comment: "Partial and section tags should be detected as standalone; trailing newline creates an empty line",
		},
		{
			name:     "Partial with complex indentation",
			template: "<div>\n    {{>partial}}\n        {{>partial}}\n</div>",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				3: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Partial tags with different indentation levels",
		},
		{
			name:     "Partial with complex indentation with trailing newline",
			template: "<div>\n    {{>partial}}\n        {{>partial}}\n</div>\n",
			want: Lines{
				0: {Type: TextLine, Segments: textContentSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				3: {Type: TextLine, Segments: textContentSegments},
				4: {Type: EmptyLine},
			},
			comment: "Partial tags with different indentation levels; trailing newline creates an empty line",
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
			name:     "Multi-block template inheritance pattern",
			template: "{{<layout}}\n{{$title}}Page Title{{/title}}\n{{$content}}\n  <p>Page content</p>\n  <ul>\n    <li>Item 1</li>\n    <li>Item 2</li>\n  </ul>\n{{/content}}\n{{/layout}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: parentBeginSegments},
				1: {Type: InlineLine, Segments: Segments{blockBeginSegment, textContentSegment, blockEndSegment}},
				2: {Type: StandaloneLine, Segments: blockBeginSegments},
				3: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				4: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				5: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				6: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				7: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				8: {Type: StandaloneLine, Segments: blockEndSegments},
				9: {Type: StandaloneLine, Segments: parentEndSegments},
			},
			comment: "Complex template inheritance with nested content blocks",
		},
		{
			name:     "Multi-block template inheritance pattern with trailing newline",
			template: "{{<layout}}\n{{$title}}Page Title{{/title}}\n{{$content}}\n  <p>Page content</p>\n  <ul>\n    <li>Item 1</li>\n    <li>Item 2</li>\n  </ul>\n{{/content}}\n{{/layout}}\n",
			want: Lines{
				0:  {Type: StandaloneLine, Segments: parentBeginSegments},
				1:  {Type: InlineLine, Segments: Segments{blockBeginSegment, textContentSegment, blockEndSegment}},
				2:  {Type: StandaloneLine, Segments: blockBeginSegments},
				3:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				4:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				5:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				6:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				7:  {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				8:  {Type: StandaloneLine, Segments: blockEndSegments},
				9:  {Type: StandaloneLine, Segments: parentEndSegments},
				10: {Type: EmptyLine},
			},
			comment: "Complex template inheritance with nested content blocks; trailing newline creates an empty line",
		},
		{
			name:     "Nested sections with indentation",
			template: "{{#outer}}\n  {{#inner}}\n    {{value}}\n  {{/inner}}\n{{/outer}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionBeginSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionEndSegment}},
				4: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			comment: "Nested sections with consistent indentation should all be standalone",
		},
		{
			name:     "Nested sections with indentation and trailing newline",
			template: "{{#outer}}\n  {{#inner}}\n    {{value}}\n  {{/inner}}\n{{/outer}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionBeginSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionEndSegment}},
				4: {Type: StandaloneLine, Segments: sectionEndSegments},
				5: {Type: EmptyLine},
			},
			comment: "Nested sections with consistent indentation should all be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Tag with dot notation and special characters",
			template: "{{user.name}}\n{{#user.settings.notifications}}\n  {{user.email}}\n{{/user.settings.notifications}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: StandaloneLine, Segments: sectionBeginSegments},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			comment: "Tags with complex paths and dot notation can be standalone",
		},
		{
			name:     "Tag with dot notation and special characters with trailing newline",
			template: "{{user.name}}\n{{#user.settings.notifications}}\n  {{user.email}}\n{{/user.settings.notifications}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: StandaloneLine, Segments: sectionBeginSegments},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				3: {Type: StandaloneLine, Segments: sectionEndSegments},
				4: {Type: EmptyLine},
			},
			comment: "Tags with complex paths and dot notation can be standalone; trailing newline creates an empty line",
		},
		{
			name:     "Comment tags",
			template: "{{! This is a comment }}\nContent\n{{! \n  Multi-line\n  comment\n}}\nMore content",
			want: Lines{
				0: {Type: StandaloneLine, Segments: commentTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: InlineLine, Segments: multilineCommentBegins},
				3: {Type: InlineLine, Segments: multilineCommentMiddles},
				4: {Type: InlineLine, Segments: multilineCommentMiddles},
				5: {Type: InlineLine, Segments: multilineCommentEnds},
				6: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Comment tags, including multi-line comments, with proper multi-line tag designation",
		},
		{
			name:     "Comment tags with trailing newline",
			template: "{{! This is a comment }}\nContent\n{{! \n  Multi-line\n  comment\n}}\nMore content\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: commentTagSegments},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: InlineLine, Segments: multilineCommentBegins},
				3: {Type: InlineLine, Segments: multilineCommentMiddles},
				4: {Type: InlineLine, Segments: multilineCommentMiddles},
				5: {Type: InlineLine, Segments: multilineCommentEnds},
				6: {Type: TextLine, Segments: textContentSegments},
				7: {Type: EmptyLine},
			},
			comment: "Comment tags, including multi-line comments, with proper multi-line tag designation; trailing newline creates an empty line",
		},
		{
			name:     "Comment tags with leading and trailing whitespace",
			template: "   {{! This is a comment }}   \nContent\n   {{! \n  Multi-line\n  comment\n}}   \nMore content",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, commentTagSegment, whitespaceSegment}},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: InlineLine, Segments: Segments{whitespaceSegment, multilineCommentBegin}},
				3: {Type: InlineLine, Segments: Segments{multilineCommentMiddle}},
				4: {Type: InlineLine, Segments: Segments{multilineCommentMiddle}},
				5: {Type: InlineLine, Segments: Segments{multilineCommentEnd, whitespaceSegment}},
				6: {Type: TextLine, Segments: textContentSegments},
			},
			comment: "Comment tags, including multi-line comments, with proper multi-line tag designation",
		},
		{
			name:     "Comment tags with leading and trailing whitespace and with trailing newline",
			template: "   {{! This is a comment }}   \nContent\n   {{! \n  Multi-line\n  comment\n}}   \nMore content\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, commentTagSegment, whitespaceSegment}},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: InlineLine, Segments: Segments{whitespaceSegment, multilineCommentBegin}},
				3: {Type: InlineLine, Segments: Segments{multilineCommentMiddle}},
				4: {Type: InlineLine, Segments: Segments{multilineCommentMiddle}},
				5: {Type: InlineLine, Segments: Segments{multilineCommentEnd, whitespaceSegment}},
				6: {Type: TextLine, Segments: textContentSegments},
				7: {Type: EmptyLine},
			},
			comment: "Comment tags, including multi-line comments, with proper multi-line tag designation; trailing newline creates an empty line",
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
			name:     "Mixed tabs and spaces with trailing newline",
			template: " \t \t{{tag}}\t \t \n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment, whitespaceSegment}},
				1: {Type: EmptyLine},
			},
			comment: "Mix of tabs and spaces counts as whitespace; trailing newline creates an empty line",
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
			name:     "Multiple sequential newlines with trailing newline",
			template: "{{tag}}\n\n\n{{tag}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
				1: {Type: EmptyLine},
				2: {Type: EmptyLine},
				3: {Type: StandaloneLine, Segments: varTagSegments},
				4: {Type: EmptyLine},
			},
			comment: "Multiple empty lines between tags; trailing newline creates an empty line",
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
		{
			name:     "Inverted Section - With content",
			template: "{{^section}}\n  Content\n{{/section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
			},
			comment: "Inverted section with content should have a standalone opening and closing line",
		},
		{
			name:     "Inverted Section - With content and trailing newline",
			template: "{{^section}}\n  Content\n{{/section}}\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginInvertedSegments},
				1: {Type: TextLine, Segments: Segments{whitespaceSegment, textContentSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndInvertedSegments},
				3: {Type: EmptyLine},
			},
			comment: "Inverted section with content should have a standalone opening and closing line and trailing newline",
		},
		{
			name:     "Custom Delimiters",
			template: "{{=<% %>=}}\n<%tag%>",
			want: Lines{
				0: {Type: StandaloneLine, Segments: delimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Custom delimiters should be recognized correctly",
		},
		{
			name:     "Custom Delimiters with trailing newline",
			template: "{{=<% %>=}}\n<%tag%>\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: delimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: EmptyLine},
			},
			comment: "Custom delimiters should be recognized correctly with trailing newline",
		},
		{
			name:     "Custom Delimiters with inline text",
			template: "{{=<% %>=}}\nText <%tag%> more text",
			want: Lines{
				0: {Type: StandaloneLine, Segments: delimiterTagSegments},
				1: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
			},
			comment: "Custom delimiters with inline text should parse correctly",
		},
		{
			name:     "Custom Delimiters with inline text",
			template: "{{=<% %>=}}\nText <%tag%> more text with trailing newline\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: delimiterTagSegments},
				1: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
				2: {Type: EmptyLine},
			},
			comment: "Custom delimiters with inline text should parse correctly with trailing newline",
		},
		{
			name:     "Unclosed Section",
			template: "{{#section}}\nContent",
			error:    "unclosed tags: [tag: {name: 'section', type: 'SectionTag'}]",
			comment:  "Unclosed sections should result in an invalid line type",
		},
		{
			name:     "Unclosed Inverted Section",
			template: "{{^section}}\nContent",
			error:    "unclosed tags: [tag: {name: 'section', type: 'InvertedSectionTag'}]",
			comment:  "Unclosed inverted sections should result in an invalid line type",
		},
		{
			name:     "Unclosed Comment",
			template: "{{! This is an unclosed comment",
			error:    "unclosed tags: [tag: {name: 'section', type: 'InvertedSectionTag'}]",
			comment:  "Unclosed comments should result in an invalid line type",
		},
		{
			name:     "Unclosed Variable",
			template: "{{var",
			error:    "invalid closing delimiter tag in '{{var'",
			comment:  "Unclosed variables should result in an invalid line type",
		},
		{
			name:     "Unclosed Triple-Brace Variable",
			template: "{{{tb_var",
			error:    "invalid closing triple-brace tag in '{{{tb_var'",
			comment:  "Unclosed triple-brace variables should result in an invalid line type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewTemplateClassifier(tt.template)
			got, err := classifier.Classify()
			switch wantError(tt.error) {
			case NO:
				switch {
				case err != nil:
					t.Error(err.Error())
				case !got.Equal(tt.want):
					showMismatch(t, got.Normalize(), tt.want.Normalize(), tt.template, tt.comment)
				}
			case YES:
				if err == nil {
					checkError(t, err, tt.error, tt.template, tt.comment)
				}
			}
		})
	}
}

const (
	YES = 'y'
	NO  = 'n'
)

func wantError(error string) byte {
	if error == "" {
		return NO
	}
	return YES
}

func checkError(t *testing.T, got error, want string, template, comment string) {
	if got == nil {
		t.Errorf("ERROR:"+
			"\n\tWant Error: %v"+
			"\n\tGot:        <No Error>"+
			"\n\tTemplate:   %q"+
			"\n\tComment:    %s\n",
			want,
			template,
			comment,
		)
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

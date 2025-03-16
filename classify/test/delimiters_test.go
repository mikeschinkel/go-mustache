package test

import (
	"testing"
)

func TestClassifierWithDelimiters(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Custom Delimiters",
			template: "{{=<% %>=}}\n<%tag%>",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			comment: "Custom delimiters should be recognized correctly",
		},
		{
			name:     "Custom Delimiters with trailing newline",
			template: "{{=<% %>=}}\n<%tag%>\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: EmptyLine},
			},
			comment: "Custom delimiters should be recognized correctly with trailing newline",
		},
		{
			name:     "Custom Delimiters with inline text",
			template: "{{=<% %>=}}\nText <%tag%> more text",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
			},
			comment: "Custom delimiters with inline text should parse correctly",
		},
		{
			name:     "Custom Delimiters with inline text",
			template: "{{=<% %>=}}\nText <%tag%> more text with trailing newline\n",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: InlineLine, Segments: Segments{textContentSegment, varTagSegment, textContentSegment}},
				2: {Type: EmptyLine},
			},
			comment: "Custom delimiters with inline text should parse correctly with trailing newline",
		},
		{
			name:     "Delimiter change with whitespace",
			template: "{{= @   @ =}}\n@tag@",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Custom delimiters with whitespace should be recognized correctly",
		},
		{
			name:     "Multiple delimiter changes",
			template: "{{=<< >>=}}\n<<var>>\n<<={{ }}=>>{{var}}\n<<var>>",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
				2: {Type: InlineLine, Segments: Segments{setDelimiterTagSegment, varTagSegment}},
				3: {Type: TextLine, Segments: textContentSegments},
			},
			error:   "",
			comment: "Multiple delimiter changes within a template",
		},
		{
			name:     "Non-standalone section tag",
			template: "start {{#section}}\nContent\n{{/section}} end",
			want: Lines{
				0: {Type: InlineLine, Segments: Segments{textContentSegment, sectionBeginSegment}},
				1: {Type: TextLine, Segments: textContentSegments},
				2: {Type: InlineLine, Segments: Segments{sectionEndSegment, textContentSegment}},
			},
			error:   "",
			comment: "Section tags with text on the same line shouldn't be considered standalone",
		},
		{
			name:     "Block tag with different identifier in closing",
			template: "{{$block1}}Content{{/block2}}",
			error:    "mismatched opening and closing delimiters; opening_identifier=block1; closing_identifier=block2; template_line={{$block1}}Content{{/block2}}",
			comment:  "Block with different identifiers in opening and closing should error",
		},
		{
			name:     "Special characters in delimiters",
			template: "{{=*~ ~*=}}\n*~tag~*",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Delimiters with special characters should be recognized correctly",
		},
		{
			name:     "Section with ampersand unescaped variable",
			template: "{{#section}}\n  {{&name}}\n{{/section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, ampersandUnescapedSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			error:   "",
			comment: "Ampersand unescaped variables inside sections should be handled correctly",
		},
		{
			name:     "Empty identifier",
			template: "{{}}",
			error:    "tag identifier is missing; template_line={{}}",
			comment:  "Empty tag identifiers should result in an error",
		},
		{
			name:     "Decimal point in identifier",
			template: "{{3.14}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Identifiers with decimal points should be valid",
		},
		{
			name:     "Mixed case template",
			template: "{{#Section}}\n  {{Name}}\n{{/Section}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, varTagSegment}},
				2: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			error:   "",
			comment: "Template identifiers with mixed case should be handled correctly",
		},
		{
			name:     "Complex nested sections with mixed content",
			template: "{{#outer}}\n  {{#inner}}\n    {{>partial}}\n    {{^inverted}}\n      {{.}}\n    {{/inverted}}\n  {{/inner}}\n{{/outer}}",
			want: Lines{
				0: {Type: StandaloneLine, Segments: sectionBeginSegments},
				1: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionBeginSegment}},
				2: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, partialTagSegment}},
				3: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionBeginInvertedSegment}},
				4: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, MakeSegment(CompleteTag, DotTag)}},
				5: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionEndInvertedSegment}},
				6: {Type: StandaloneLine, Segments: Segments{whitespaceSegment, sectionEndSegment}},
				7: {Type: StandaloneLine, Segments: sectionEndSegments},
			},
			error:   "",
			comment: "Complex nested sections with different tag types should work together",
		},
		{
			name:     "Delimiter change with whitespace around",
			template: "{{= @@ @@  =}}\n@@tag@@",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Whitespace around the delimiter definitions should be ignored",
		},
		{
			name:     "Delimiter change to long and funky delimiter characters",
			template: "{{=~<{[ ]}>~=}}\n~<{[tag]}>~",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Long delimiter with funky characters are valid",
		},
		{
			name:     "Triple brace delimiter rejection",
			template: "{{={{{ }}}=}}\n{{{tag}}}",
			error:    "triple braces cannot be used as custom delimiters; delimiter_position=opening_delimiter; template_line={{={{{ }}}=}}",
			comment:  "Setting delimiters to triple braces should be rejected to avoid ambiguity",
		},
		{
			name:     "Single character delimiters",
			template: "{{=| |=}}\n|tag|",
			want: Lines{
				0: {Type: StandaloneLine, Segments: setDelimiterTagSegments},
				1: {Type: StandaloneLine, Segments: varTagSegments},
			},
			error:   "",
			comment: "Single character delimiters should be recognized correctly",
		},
		{
			name:     "Empty delimiters",
			template: "{{==}}",
			error:    "delimiters cannot be empty; template_line={{==}}",
			comment:  "Empty delimiters should result in an error",
		},
	})
}

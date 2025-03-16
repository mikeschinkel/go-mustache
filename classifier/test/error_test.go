package test

import (
	"testing"
)

func TestClassifierWithErrorScenarios(t *testing.T) {
	testsRunner(t, []TestCase{
		{
			name:     "Unclosed Section",
			template: "{{#section}}\nContent",
			error:    "unclosed tags\ntag_stack=[tag: {name: 'section', type: 'SectionTag'}]",
			comment:  "Unclosed sections should result in an invalid line type",
		},
		{
			name:     "Unclosed Inverted Section",
			template: "{{^section}}\nContent",
			error:    "unclosed tags\ntag_stack=[tag: {name: 'section', type: 'InvertedSectionTag'}]",
			comment:  "Unclosed inverted sections should result in an invalid line type",
		},
		{
			name:     "Unclosed Comment",
			template: "{{! This is an unclosed comment",
			error:    "unclosed comment tag; template_content={{! This is an unclosed comment",
			comment:  "Unclosed comments should result in an invalid line type",
		},
		{
			name:     "Unclosed Variable",
			template: "{{var",
			error:    "invalid closing tag; delimiter_type=standard_delimiter; template_line={{var",
			comment:  "Unclosed variables should result in an invalid line type",
		},
		{
			name:     "Unclosed Triple-Brace Variable",
			template: "{{{tb_var",
			error:    "invalid closing tag; delimiter_type=triple_brace_delimiter; template_line={{{tb_var",
			comment:  "Unclosed triple-brace variables should result in an invalid line type",
		},
		{
			name:     "Unbalanced closing section",
			template: "Text\n{{/section}}\nMore text",
			error:    "mismatched opening and closing delimiters; opening_identifier=; closing_identifier=section; template_line={{/section}}",
			comment:  "Unbalanced closing section without matching opening should error",
		},
		{
			name:     "Nested mismatched sections",
			template: "{{#outer}}{{#inner}}content{{/outer}}{{/inner}}",
			error:    "mismatched opening and closing delimiters; opening_identifier=outer; closing_identifier=inner; template_line={{#outer}}{{#inner}}content{{/outer}}{{/inner}}",
			comment:  "Nested sections with mismatched closing tags should error",
		},
	})
}

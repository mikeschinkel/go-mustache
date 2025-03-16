package test

import (
	"testing"
)

func TestClassifierWithErrorScenarios(t *testing.T) {
	testsRunner(t, []TestCase{
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
			error:    "invalid unclosed comment tag in '{{! This is an unclosed comment'",
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
		{
			name:     "Unbalanced closing section",
			template: "Text\n{{/section}}\nMore text",
			error:    "mismatched opening and closing delimiters (''!='section') in '{{/section}}'",
			comment:  "Unbalanced closing section without matching opening should error",
		},
		{
			name:     "Nested mismatched sections",
			template: "{{#outer}}{{#inner}}content{{/outer}}{{/inner}}",
			error:    "mismatched opening and closing delimiters ('outer'!='inner') in '{{#outer}}{{#inner}}content{{/outer}}{{/inner}}'",
			comment:  "Nested sections with mismatched closing tags should error",
		},
	})
}

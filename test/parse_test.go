// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"reflect"
	"testing"

	"github.com/alexkappa/mustache"
)

func TestParser(t *testing.T) {
	for _, test := range []struct {
		template string
		expected []Node
	}{
		{
			"{{#foo}}\n\t{{#foo}}hello nested{{/foo}}{{/foo}}",
			[]Node{
				&SectionNode{Name: "foo", Elems: []Node{
					TextNode("\n\t"),
					&SectionNode{Name: "foo", Elems: []Node{
						TextNode("hello nested"),
					}},
				}},
			},
		},
		{
			"\nfoo {{bar}} {{#alex}}\r\n\tbaz\n{{/alex}} {{!foo}}",
			[]Node{
				TextNode("\nfoo "),
				&VarNode{Name: "bar", Escape: true},
				TextNode(" "),
				&SectionNode{Name: "alex", Elems: []Node{
					TextNode("\r\n\tbaz\n"),
				}},
				TextNode(" "),
				CommentNode("foo"),
			},
		},
		{
			"this will{{^foo}}not{{/foo}} be rendered",
			[]Node{
				TextNode("this will"),
				&SectionNode{Name: "foo", Inverted: true, Elems: []Node{
					TextNode("not"),
				}},
				TextNode(" be rendered"),
			},
		},
		{
			"{{#list}}({{.}}){{/list}}",
			[]Node{
				&SectionNode{Name: "list", Elems: []Node{
					TextNode("("),
					&VarNode{Name: ".", Escape: true},
					TextNode(")"),
				}},
			},
		},
	} {
		parser := mustache.NewParser(mustache.NewLexer(test.template, "{{", "}}"))
		elems, err := parser.Parse()
		if err != nil {
			t.Fatal(err)
		}
		for i, elem := range elems {
			if !reflect.DeepEqual(elem, test.expected[i]) {
				t.Errorf("elements are not equal %v != %v", elem, test.expected[i])
			}
		}
	}
}

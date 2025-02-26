// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/alexkappa/mustache"
)

//goland:noinspection GoUnusedExportedType
type (
	Node        = mustache.Node
	TextNode    = mustache.TextNode
	CommentNode = mustache.CommentNode
	VarNode     = mustache.VarNode
	SectionNode = mustache.SectionNode
	PartialNode = mustache.PartialNode
	DelimNode   = mustache.DelimNode
)

var tests = []struct {
	name     string
	mustache string
	want     string
	context  any
}{
	{
		name:     "simple mustache",
		mustache: "some text {{foo}} here",
		want:     "some text bar here",
		context:  `{"foo": "bar"}`,
	},
	{
		name:     "nested mustache",
		mustache: "{{#foo}} foo is defined {{bar}} {{/foo}}",
		want:     " foo is defined baz ",
		context:  `{"foo": {"bar": "baz"}}`,
	},
	{
		name:     "Various variable resolutions",
		mustache: "{{#basics}}{{#profiles}}|{{network}}{{#network}}|{{network}}|{{.}}{{/network}}{{/profiles}}|{{/basics}}",
		want:     "|GitHub|GitHub|GitHub|LinkedIn|LinkedIn|LinkedIn|Twitter|Twitter|Twitter|",
		context:  `{"basics":{"profiles":[{"network":"GitHub"},{"network":"LinkedIn"},{"network":"Twitter"}]}}`,
	},
}

func TestTemplate(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.mustache)
			m := mustache.New()
			err := m.Parse(r)
			if err != nil {
				t.Errorf("ERROR: Failed to parse; %s", err.Error())
			}
			// Context can be a Go value such as a map, slice or struct
			context := tt.context
			if reflect.TypeOf(context).Kind() == reflect.String {
				// But if a string, expect JSON and unmarshal to a Go value
				err = json.Unmarshal([]byte(tt.context.(string)), &context)
				if err != nil {
					t.Errorf("ERROR: Failed to unmarshal JSON context; %s", err.Error())
				}
			}

			var got bytes.Buffer
			err = m.Render(&got, context)
			if err != nil {
				t.Errorf("ERROR: Failed to render; %s", err.Error())
			}
			if got.String() != tt.want {
				t.Errorf("ERROR: Wanted %q got %q", tt.want, got.String())
			}
		})
	}
}

func TestFalsyTemplate(t *testing.T) {
	input := strings.NewReader("some text {{^foo}}{{foo}}{{/foo}} {{bar}} here")
	template := mustache.New()
	err := template.Parse(input)
	if err != nil {
		t.Error(err)
	}
	buf := bytes.Buffer{}
	output := mustache.NewWriter(&buf)
	err = template.Render(output, map[string]interface{}{"foo": 0, "bar": false})
	if err != nil {
		t.Error(err)
	}
	want := "some text 0 false here"
	if output.String() != want {
		t.Errorf("want %q got %q", want, output.String())
	}
}

func TestParseTree(t *testing.T) {
	template := mustache.New()
	template.Elems = []mustache.Node{
		TextNode("Lorem ipsum dolor sit "),
		&VarNode{Name: "foo", Escape: false},
		TextNode(", "),
		&SectionNode{
			Name:     "bar",
			Inverted: false,
			Elems: []Node{
				&VarNode{Name: "baz", Escape: true},
				TextNode(" adipiscing"),
			}},
		TextNode(" elit. Proin commodo viverra elit "),
		&VarNode{Name: "zer", Escape: false},
		TextNode("."),
	}
	data := map[string]interface{}{
		"foo": "amet",
		"bar": map[string]string{"baz": "consectetur"},
		"zer": 0.11,
	}
	b := bytes.NewBuffer(nil)
	w := mustache.NewWriter(b)
	for _, e := range template.Elems {
		err := e.Render(template, w, data)
		if err != nil {
			t.Error(err)
		}
	}
	must("flush", w.Flush())

	want := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Proin commodo viverra elit 0.11.`
	if want != b.String() {
		t.Errorf("want didn't match. want %q got %q.", want, b.String())
		t.Log(b.String())
	}
}

func must(what string, err error) {
	if err != nil {
		log.Printf("ERROR: Failed to %s; %s\n", what, err.Error())
	}
}

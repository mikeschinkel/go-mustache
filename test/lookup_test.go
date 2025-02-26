// Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"reflect"
	"testing"

	"github.com/alexkappa/mustache"
)

type Bugs struct{}

func (b *Bugs) Bunny() string {
	return "What's up, Doc!"
}

func TestSimpleLookup(t *testing.T) {
	for _, test := range []struct {
		context    interface{}
		assertions []struct {
			name  string
			value interface{}
			truth bool
		}
	}{
		{
			context: map[string]interface{}{
				"integer": 123,
				"string":  "abc",
				"boolean": true,
				"map": map[string]interface{}{
					"in": "I'm nested!",
				},
				"ptr": &struct {
					Foo *struct{ Bar string }
				}{
					Foo: &struct{ Bar string }{
						Bar: "bar",
					},
				},
			},
			assertions: []struct {
				name  string
				value interface{}
				truth bool
			}{
				{"integer", 123, true},
				{"string", "abc", true},
				{"boolean", true, true},
				{"map.in", "I'm nested!", true},
				{"ptr.Foo.Bar", "bar", true},
			},
		},
		{
			context: struct {
				Integer int
				String  string
				Boolean bool
				Nested  struct{ Inside string }
			}{
				123, "abc", true, struct{ Inside string }{"I'm nested!"},
			},
			assertions: []struct {
				name  string
				value interface{}
				truth bool
			}{
				{"Integer", 123, true},
				{"String", "abc", true},
				{"Boolean", true, true},
				{"Nested.Inside", "I'm nested!", true},
			},
		},
		{
			context: struct {
				Integer int    `template:"int"`
				String  string `template:"str"`
				Boolean bool   `template:"bool"`
				Nested  struct {
					Inside string `template:"inside"`
				} `template:"nested"`
			}{
				Integer: 123,
				String:  "abc",
				Boolean: true,
				Nested: struct {
					Inside string `template:"inside"`
				}{"I'm nested!"},
			},
			assertions: []struct {
				name  string
				value interface{}
				truth bool
			}{
				{"int", 123, true},
				{"str", "abc", true},
				{"bool", true, true},
				{"nested.inside", "I'm nested!", true},
			},
		},
		{
			context: Bugs{},
			assertions: []struct {
				name  string
				value interface{}
				truth bool
			}{
				{name: "Bunny", value: "What's up, Doc!", truth: true},
			},
		},
	} {
		for _, assertion := range test.assertions {
			tmpl := mustache.New(mustache.StructTag("template"))
			value, truth, err := tmpl.Lookup(assertion.name, test.context)
			if err != nil {
				t.Errorf("Unexpected error %v != %v", value, err)
			}
			if value != assertion.value {
				t.Errorf("Unexpected value %v != %v", value, assertion.value)
			}
			if truth != assertion.truth {
				t.Errorf("Unexpected truth %t != %t", truth, assertion.truth)
			}
		}
	}
}

func TestTruth(t *testing.T) {
	for _, test := range []struct {
		input    interface{}
		expected bool
	}{
		{"abc", true},
		{"", false},
		{123, true},
		{0, false},
		{true, true},
		{false, false},
	} {
		truth := mustache.Truthiness(reflect.ValueOf(test.input))
		if truth != test.expected {
			t.Errorf("Unexpected truth %t != %t", truth, test.expected)
		}
	}
}

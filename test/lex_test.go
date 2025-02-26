// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"testing"

	"github.com/alexkappa/mustache"
)

type (
	Token = mustache.Token
)

const (
	TokenText       = mustache.TokenText
	TokenLeftDelim  = mustache.TokenLeftDelim
	TokenRawStart   = mustache.TokenRawStart
	TokenIdentifier = mustache.TokenIdentifier
	TokenRawEnd     = mustache.TokenRawEnd
	TokenRightDelim = mustache.TokenRightDelim
	TokenComment    = mustache.TokenComment
	TokenSetDelim   = mustache.TokenSetDelim
	TokenEOF        = mustache.TokenEOF
)

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		template string
		expected []Token
	}{
		{
			"foo {{{bar}}}\nbaz {{! this is ignored }}",
			[]Token{
				{Type: TokenText, Value: "foo "},
				{Type: TokenLeftDelim, Value: "{{"},
				{Type: TokenRawStart, Value: "{"},
				{Type: TokenIdentifier, Value: "bar"},
				{Type: TokenRawEnd, Value: "}"},
				{Type: TokenRightDelim, Value: "}}"},
				{Type: TokenText, Value: "\nbaz "},
				{Type: TokenLeftDelim, Value: "{{"},
				{Type: TokenComment, Value: "!"},
				{Type: TokenText, Value: " this is ignored "},
				{Type: TokenRightDelim, Value: "}}"},
				{Type: TokenEOF},
			},
		},
		{
			"\nfoo {{bar}} baz {{=| |=}}\r\n |foo| |={{ }}=| {{bar}}",
			[]Token{
				{Type: TokenText, Value: "\nfoo "},
				{Type: TokenLeftDelim, Value: "{{"},
				{Type: TokenIdentifier, Value: "bar"},
				{Type: TokenRightDelim, Value: "}}"},
				{Type: TokenText, Value: " baz "},
				{Type: TokenSetDelim},
				{Type: TokenText, Value: "\r\n "},
				{Type: TokenLeftDelim, Value: "|"},
				{Type: TokenIdentifier, Value: "foo"},
				{Type: TokenRightDelim, Value: "|"},
				{Type: TokenText, Value: " "},
				{Type: TokenSetDelim},
				{Type: TokenText, Value: " "},
				{Type: TokenLeftDelim, Value: "{{"},
				{Type: TokenIdentifier, Value: "bar"},
				{Type: TokenRightDelim, Value: "}}"},
				{Type: TokenEOF},
			},
		},
	} {
		var (
			lexer = mustache.NewLexer(test.template, "{{", "}}")
			token = lexer.Token()
			i     = 0
		)
		for token.Type > TokenEOF {
			//t.Logf("%s\n", token)
			if i >= len(test.expected) {
				t.Fatalf("token stream exceeded the length of expected Tokens.")
			}
			if token.Type != test.expected[i].Type {
				t.Errorf("unexpected token %q, expected %q", token.Type, test.expected[i].Type)
			}
			if token.Value != test.expected[i].Value {
				t.Errorf("unexpected value %q, expected %q", token.Value, test.expected[i].Value)
			}
			token = lexer.Token()
			i++
		}
	}
}

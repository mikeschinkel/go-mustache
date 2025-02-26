// Copyright (c) 2014 Alex Kalyvitis
// Portions Copyright (c) 2011 The Go Authors

package mustache

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Token represents a token or text string returned from the scanner.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// String satisfies the fmt.Stringer interface making it easier to print Tokens.
func (i Token) String() string {
	return fmt.Sprintf("%s:%q", i.Type, i.Value)
}

// TokenType identifies the type of lex Tokens.
type TokenType int

const (
	TokenError TokenType = iota // error occurred; value is text of error
	TokenEOF
	TokenIdentifier        // alphanumeric identifier
	TokenLeftDelim         // {{ left action delimiter
	TokenRightDelim        // }} right action delimiter
	TokenText              // plain text
	TokenComment           // {{! this is a comment and is ignored}}
	TokenSectionStart      // {{#foo}} denotes a section start
	TokenSectionInverse    // {{^foo}} denotes an inverse section start
	TokenSectionEnd        // {{/foo}} denotes the closing of a section
	TokenRawStart          // { denotes the beginning of an unencoded identifier
	TokenRawEnd            // } denotes the end of an unencoded identifier
	TokenRawAlt            // {{&foo}} is an alternative way to define raw tags
	TokenPartial           // {{>foo}} denotes a partial
	TokenSetDelim          // {{={% %}=}} sets delimiters to {% and %}
	TokenSetLeftDelim      // denotes a custom left delimiter
	TokenSetRightDelim     // denotes a custom right delimiter
	TokenDynamicStart      // {{*foo}} denotes a dynamic name lookup for partial foo
	TokenDot               // {{*foo.*bar}} denotes a dotted dynamic name lookup for partial foo
	TokenLeadingWhitespace // whitespace at start of line
	TokenNewlineText       // text containing only a newline
)

// TokenName used for pretty printing types.
var TokenName = map[TokenType]string{
	TokenError:             "t_error",
	TokenEOF:               "t_eof",
	TokenIdentifier:        "t_ident",
	TokenLeftDelim:         "t_left_delim",
	TokenRightDelim:        "t_right_delim",
	TokenText:              "t_text",
	TokenComment:           "t_comment",
	TokenSectionStart:      "t_section_start",
	TokenSectionInverse:    "t_section_inverse",
	TokenSectionEnd:        "t_section_end",
	TokenRawStart:          "t_raw_start",
	TokenRawEnd:            "t_raw_end",
	TokenRawAlt:            "t_raw_alt",
	TokenPartial:           "t_partial",
	TokenSetDelim:          "t_set_delim",
	TokenSetLeftDelim:      "t_set_left_delim",
	TokenSetRightDelim:     "t_set_right_delim",
	TokenDynamicStart:      "t_dynamic_start",
	TokenDot:               "t_dot",
	TokenLeadingWhitespace: "t_leading_whitespace",
	TokenNewlineText:       "t_newline_text",
}

// String satisfies the fmt.Stringer interface making it easier to print Tokens.
func (i TokenType) String() string {
	s := TokenName[i]
	if s == "" {
		return fmt.Sprintf("t_unknown_%d", int(i))
	}
	return s
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the
// next state.
type stateFn func(*Lexer) stateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	name       string     // the name of the input; used only for error reports.
	input      string     // the string being scanned.
	leftDelim  string     // start of action.
	rightDelim string     // end of action.
	state      stateFn    // the next lexing function to enter.
	pos        int        // current position in the input.
	start      int        // start position of this Token.
	width      int        // width of last rune read from input.
	Tokens     chan Token // channel of scanned Tokens.
}

func (l *Lexer) matchesLeftDelim() bool {
	return strings.HasPrefix(l.input[l.pos:], l.leftDelim)
}

// next returns the next rune in the input.
func (l *Lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// seek advances the pointer by n spaces.
func (l *Lexer) seek(n int) {
	l.pos += n
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// emit passes an token back to the client.
func (l *Lexer) emit(t TokenType) {
	l.Tokens <- Token{
		t,
		l.input[l.start:l.pos],
		l.lineNum(),
		l.columnNum(),
	}
	l.start = l.pos
}

// emit passes an token back to the client using the starting and ending positions
func (l *Lexer) emitFor(t TokenType, startPos, endPos int) {
	startSave, endSave := l.start, l.pos
	l.start, l.pos = startPos, endPos
	l.emit(t)
	l.start, l.pos = startSave, endSave
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// char returns byte from .input at current .pos
func (l *Lexer) char() byte {
	return l.input[l.pos]
}

// maybeEmitIndent emits a TokenLeadingWhitespace with a whitespace indention string
// prefixing a partial, if applicable.
func (l *Lexer) maybeEmitIndent() {
	var startPos, endPos int
	if l.pos == 0 {
		goto end
	}
	endPos = l.pos
	for startPos = l.pos - 1; startPos > 0; startPos-- {
		if !unicode.IsSpace(rune(l.input[startPos])) {
			goto end
		}
		if l.input[startPos] == '\n' {
			// Omit the \n by adding 1 back to startPos
			startPos++
			// Now break out and
			break
		}
	}
	l.emitFor(TokenLeadingWhitespace, startPos, endPos)
end:
}

// lineNum reports which Line we're on. Doing it this way
// means we don't have to worry about peek double counting.
func (l *Lexer) lineNum() int {
	return 1 + strings.Count(l.input[:l.pos], "\n")
}

// columnNum reports the character of the current Line we're on.
func (l *Lexer) columnNum() int {
	if lf := strings.LastIndex(l.input[:l.pos], "\n"); lf != -1 {
		return len(l.input[lf+1 : l.pos])
	}
	return len(l.input[:l.pos])
}

// error returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.Token.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.Tokens <- Token{
		TokenError,
		fmt.Sprintf(format, args...),
		l.lineNum(),
		l.columnNum(),
	}
	return nil
}

// Token returns the next token from the input.
func (l *Lexer) Token() Token {
	for {
		select {
		case token := <-l.Tokens:
			return token
		default:
			l.state = l.state(l)
		}
	}
}

func (l *Lexer) String() string {
	w := bytes.NewBuffer(nil)
	MustFprintf(w, "Template: %q\n", l.input)
	MustFprintf(w, "Index   : %q\n", l.pos)
	MustFprintf(w, "Current : %q\n", l.char())
	MustFprintf(w, "Buffer  : %q\n", l.input[l.start:l.pos])
	return w.String()
}

// NewLexer creates a new scanner for the input string.
func NewLexer(input, left, right string) *Lexer {
	l := &Lexer{
		input:      input,
		leftDelim:  left,
		rightDelim: right,
		Tokens:     make(chan Token, 2),
	}
	l.state = stateText // initial state
	return l
}

// state functions.

// stateText scans until an opening action delimiter, "{{".
func stateText(l *Lexer) (fn stateFn) {
	for {
		// Lookahead for {{ which should switch to lexing an open tag instead of
		// regular text Tokens.
		if l.matchesLeftDelim() {
			if l.pos > l.start {
				l.emit(TokenText)
				l.maybeEmitIndent()
			}
			fn = stateLeftDelim
			goto end
		}
		// Produce a Token and exit the loop if we have reached the end of file.
		if l.next() == eof {
			break
		}
	}
	// Emit whatever we gathered so far as text.
	if l.pos > l.start {
		l.emit(TokenText)
	}
	// Always end with EOF Token. The parser will keep asking for Tokens until
	// an TokenEOF or TokenError Token are encountered.
	l.emit(TokenEOF)
	// The text state doesn't have a default next state.
end:
	return fn
}

// stateLeftDelim scans the left delimiter, which is known to be present.
func stateLeftDelim(l *Lexer) stateFn {
	l.seek(len(l.leftDelim))
	if l.peek() == '=' {
		// When the Lexer encounters "{{=" it proceeds to the set delimiter
		// state which alters the left and right delimiters. This operation is
		// hidden from the parser and no Tokens are emitted.
		l.next()
		return stateSetDelim
	}
	l.emit(TokenLeftDelim)
	return stateTag
}

// stateRightDelim scans the right delimiter, which is known to be present.
func stateRightDelim(l *Lexer) stateFn {
	l.seek(len(l.rightDelim))
	l.emit(TokenRightDelim)
	return stateText
}

// stateTag scans the elements inside action delimiters.
func stateTag(l *Lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], "}"+l.rightDelim) {
		l.seek(1)
		l.emit(TokenRawEnd)
		return stateRightDelim
	}
	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
		return stateRightDelim
	}
	switch r := l.next(); {
	case r == eof || r == '\n':
		return l.errorf("unclosed action")
	case whitespace(r):
		l.ignore()
	case r == '!':
		l.emit(TokenComment)
		return stateComment
	case r == '#':
		l.emit(TokenSectionStart)
	case r == '*':
		l.emit(TokenDynamicStart)
	case r == '^':
		l.emit(TokenSectionInverse)
	case r == '/':
		l.emit(TokenSectionEnd)
	case r == '&':
		l.emit(TokenRawAlt)
	case r == '>':
		l.emit(TokenPartial)
	case r == '{':
		l.emit(TokenRawStart)
	case r == '.':
		l.emit(TokenDot)
	case alphanum(r):
		l.backup()
		return stateIdent
	default:
		return l.errorf("unrecognized character in action: %#U", r)
	}
	return stateTag
}

// stateIdent scans an alphanumeric or field.
func stateIdent(l *Lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '.':
			l.backup()
			l.emit(TokenIdentifier)
			l.next()
			l.emit(TokenDot)
			goto end
		case alphanum(r):
			// absorb
		default:
			l.backup()
			l.emit(TokenIdentifier)
			goto end
		}
	}
end:
	return stateTag
}

// stateComment scans a comment. The left comment marker is known to be present.
func stateComment(l *Lexer) stateFn {
	i := strings.Index(l.input[l.pos:], l.rightDelim)
	if i < 0 {
		return l.errorf("unclosed tag")
	}
	l.seek(i)
	l.emit(TokenText)
	return stateRightDelim
}

// stateSetDelim scans a set of set delimiter tags and replaces the lexers left
// and right delimiters to new values.
func stateSetDelim(l *Lexer) stateFn {
	end := "=" + l.rightDelim
	i := strings.Index(l.input[l.pos:], end)
	if i < 0 {
		return l.errorf("unclosed tag")
	}
	delims := strings.Split(l.input[l.pos:l.pos+i], " ") // " | | "
	if len(delims) < 2 {
		l.errorf("set delimiters should be separated by a space")
	}
	delimFn := leftFn
	for _, delim := range delims {
		if delim != "" {
			if delimFn != nil {
				delimFn = delimFn(l, delim)
			}
		}
	}
	l.seek(i + len(end))
	l.ignore()
	l.emit(TokenSetDelim)
	return stateText
}

// delimFn is a self referencing function which helps with setting the right
// delimiter in the right order.
type delimFn func(l *Lexer, s string) delimFn

// leftFn sets the left delimiter to s and returns a rightFn.
func leftFn(l *Lexer, s string) delimFn {
	l.leftDelim = s
	return rightFn
}

// rightFn sets the right delimiter to s.
func rightFn(l *Lexer, s string) delimFn {
	l.rightDelim = s
	return nil
}

// whitespace reports whether r is a space character.
func whitespace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	}
	return false
}

// alphanum reports whether r is an alphabetic, digit, or underscore.
func alphanum(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

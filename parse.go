// Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"fmt"
	"io"
	"strings"
)

type Parser struct {
	lexer      *Lexer
	buf        []Token
	ast        []Node
	silentMiss bool
}

// read returns the next token from the lexer and advances the cursor. This
// token will not be available by the parser after it has been read.
func (p *Parser) read() Token {
	if len(p.buf) > 0 {
		r := p.buf[0]
		p.buf = p.buf[1:]
		return r
	}
	return p.lexer.Token()
}

// readn returns the next n Tokens from the lexer and advances the cursor. If it
// couldn't read all n Tokens, for example if a TokenEOF was returned by the
// lexer, an error is returned and the returned slice will have all Tokens read
// until that point, including TokenEOF.
func (p *Parser) readn(n int) ([]Token, error) {
	Tokens := make([]Token, 0, n) // make a slice capable of storing up to n Tokens
	for i := 0; i < n; i++ {
		Tokens = append(Tokens, p.read())
		if Tokens[i].Type == TokenEOF {
			return Tokens, io.EOF
		}
	}
	return Tokens, nil
}

// readt returns the Tokens starting from the current position until the first
// match of t. Similar to readn it will return an error if a TokenEOF was
// returned by the lexer before a match was made.
func (p *Parser) readt(t TokenType) ([]Token, error) {
	var tokens []Token
	for {
		token := p.read()
		tokens = append(tokens, token)
		//goland:noinspection GoSwitchMissingCasesForIotaConsts
		switch token.Type {
		case TokenEOF:
			return tokens, fmt.Errorf("token %q not found", t)
		case t:
			return tokens, nil
		}
	}
}

// readv returns the Tokens starting from the current position until the first
// match of t. A match is made only of t.Type and t.Value are equal to the examined
// Token.
func (p *Parser) readv(t Token) ([]Token, error) {
	var tokens []Token
	for {
		read, err := p.readt(t.Type)
		tokens = append(tokens, read...)
		if err != nil {
			return tokens, err
		}
		if len(read) > 0 && read[len(read)-1].Value == t.Value {
			break
		}
	}
	return tokens, nil
}

// peek returns the next token without advancing the cursor. Consecutive calls
// of peek would result in the same token being retuned. To advance the cursor,
// a read must be made.
func (p *Parser) peek() Token {
	if len(p.buf) > 0 {
		return p.buf[0]
	}
	t := p.lexer.Token()
	p.buf = append(p.buf, t)
	return t
}

// peekn returns the next n Tokens without advancing the cursor.
func (p *Parser) peekn(n int) ([]Token, error) {
	if len(p.buf) > n {
		return p.buf[:n], nil
	}
	for i := len(p.buf) - 1; i < n; i++ {
		t := p.lexer.Token()
		p.buf = append(p.buf, t)
		if t.Type == TokenEOF {
			return p.buf, io.EOF
		}
	}
	return p.buf, nil
}

// peekt returns the Tokens from the current position until the first match of
// t. it will not advance the cursor.
func (p *Parser) peekt(t TokenType) ([]Token, error) {
	for i := 0; i < len(p.buf); i++ {
		//goland:noinspection GoSwitchMissingCasesForIotaConsts
		switch p.buf[i].Type {
		case t:
			return p.buf[:i], nil
		case TokenEOF:
			return p.buf[:i], io.EOF
		}
	}
	for {
		token := p.lexer.Token()
		p.buf = append(p.buf, token)
		//goland:noinspection GoSwitchMissingCasesForIotaConsts
		switch token.Type {
		case t:
			return p.buf, nil
		case TokenEOF:
			return p.buf, io.EOF
		}
	}
}

func (p *Parser) errorf(t Token, format string, v ...interface{}) error {
	if strings.HasPrefix(format, "unexpected token") {
		noop()
	}
	return fmt.Errorf("%d:%d syntax error: %s", t.Line, t.Column, fmt.Sprintf(format, v...))
}

// Parse begins parsing based on Tokens read from the lexer.
func (p *Parser) Parse() (nodes []Node, err error) {
	var node Node
	for {
		token := p.read()
		switch token.Type {
		case TokenEOF:
			goto end
		case TokenError:
			err = p.errorf(token, "%s", token.Value)
			goto end
		case TokenText:
			nodes = append(nodes, TextNode(token.Value))
		case TokenLeadingWhitespace:
			nodes = append(nodes, LeadingWhitespaceNode(token.Value))
		case TokenLeftDelim:
			node, err = p.parseTag()
			if err != nil {
				goto end
			}
			nodes = append(nodes, node)
		case TokenRawStart:
			node, err = p.parseRawTag()
			if err != nil {
				goto end
			}
			nodes = append(nodes, node)
		case TokenSetDelim:
			nodes = append(nodes, new(DelimNode))
		default:
			print()
		}
	}
end:
	return nodes, err
}

// parseTag parses a beginning of a mustache tag. It is assumed that a leftDelim
// was already read by the parser.
func (p *Parser) parseTag() (node Node, err error) {
	token := p.read()
	switch token.Type {
	case TokenIdentifier:
		node, err = p.parseVar(token, true)
	case TokenRawStart:
		node, err = p.parseRawTag()
	case TokenRawAlt:
		node, err = p.parseVar(p.read(), false)
	case TokenComment:
		node, err = p.parseComment()
	case TokenSectionInverse:
		node, err = p.parseSection(true)
	case TokenSectionStart:
		node, err = p.parseSection(false)
	case TokenPartial:
		node, err = p.parsePartial()
	case TokenDot:
		node, err = p.parseDot()
	default:
		err = p.errorf(token, "unexpected token '%s'", token)
	}
	return node, err
}

// parseDot handles a standalone dot token, which represents the current context
// in a regular mustache tag like {{.}}.
//
// This function expects to find a right delimiter (TokenRightDelim) immediately
// following the dot token. If any other token is encountered, an error is returned.
//
// Returns:
//   - A VarNode with Name="." and Escape=false
//   - An error if the right delimiter is not found after the dot
func (p *Parser) parseDot() (n Node, err error) {
	t := p.read()
	// Expect closing delimiter
	if t.Type != TokenRightDelim {
		err = p.errorf(t, "unexpected token %s; expected %s or %s",
			t, TokenRightDelim, TokenRawEnd,
		)
		goto end
	}
	// Return a VarNode with the special "." name
	n = &VarNode{
		Name:   ".",
		Escape: true,
	}
end:
	return n, err
}

// parsePath processes an identifier path that is potentially dot-separated by
// starting with an initial identifier. It extends the provided initial path
// string by parsing any subsequent dot-notation components (e.g.,
// ".field1.field2") that follow in the token stream.
//
// For example, when processing "person.name.first", if "person" was already consumed
// and passed as the initial path, this function would append ".name.first" to create
// the complete path "person.name.first".
//
// This function continues parsing until it encounters a non-dot token, at which
// point it stops without consuming that token, leaving it for the caller to
// process. So if there is no dot in the path it simply returns the same path
// that was passed to it.
//
// Parameters:
//   - path: The initial path (usually identifier that's already been consumed)
//
// Returns:
//   - The complete path string with all dot components appended
//   - An error if a dot is not followed by a valid identifier
func (p *Parser) parsePath(path string) (_ string, err error) {
	var next Token

	// Check for dotted path components
	for {
		next = p.peek()
		if next.Type != TokenDot {
			break
		}
		// Consume the dot
		p.read()

		// The next token should be an identifier
		next = p.read()
		if next.Type != TokenIdentifier {
			err = p.errorf(next, "expected identifier after dot, got %s", next)
			goto end
		}

		// Add the dot and path component to the full path
		path += "." + next.Value
	}
end:
	return path, err
}

// parseRawTag parses a simple variable tag, raw (triple mustache) variable tag
// which may include dotted paths, or a standalone dot referring to the current
// context.
// For example:
//
//   - {{{name}}}           (simple identifier)
//   - {{{person.name}}}    (dotted path)
//   - {{{.}}}              (current context)
func (p *Parser) parseRawTag() (n Node, err error) {
	// Read the content token
	t := p.read()

	// Process the content token
	var name string
	switch t.Type {
	case TokenIdentifier:
		// Process identifier (possibly with path)
		name, err = p.parsePath(t.Value)
		if err != nil {
			goto end
		}
	case TokenDot:
		// Use dot as name
		name = "."
	default:
		goto end
	}

	// Check for proper closing tokens (common for all content types)
	t = p.read()
	if t.Type != TokenRawEnd {
		goto end
	}

	t = p.read()
	if t.Type != TokenRightDelim {
		goto end
	}

	n = &VarNode{Name: name, Escape: false}

end:
	if n == nil {
		err = p.errorf(t, "unexpected token %s", t)
	}
	return n, err
}

// parseVar parses a variable tag, which may include dotted paths.
// For example: {{person.name}}
func (p *Parser) parseVar(ident Token, escape bool) (n Node, err error) {
	var t Token
	var path string

	path, err = p.parsePath(ident.Value)
	if err != nil {
		goto end
	}

	// Expect the closing delimiter
	t = p.read()
	if t.Type != TokenRightDelim {
		err = p.errorf(t, "unexpected token %s", t)
		goto end
	}

	n = &VarNode{
		Name:   path,
		Escape: escape,
	}
end:
	return n, err
}

// parseComment parses a comment block. It is assumed that the next read should
// return a t_comment token.
func (p *Parser) parseComment() (node Node, err error) {
	var comment string
	for {
		t := p.read()
		switch t.Type {
		case TokenEOF:
			err = p.errorf(t, "unexpected token %s", t)
			goto end
		case TokenError:
			err = p.errorf(t, "token error %s", t.Value)
			goto end
		case TokenRightDelim:
			node = CommentNode(comment)
			goto end
		default:
			comment += t.Value
		}
	}
end:
	return node, err
}

// parseSection parses a section block. It is assumed that the next read should
// return a t_section token.`
func (p *Parser) parseSection(inverse bool) (section Node, err error) {
	var (
		nodes        []Node
		read, tokens []Token
		next         Token
		stack        = 1
	)
	t := p.read()
	if t.Type != TokenIdentifier {
		err = p.errorf(t, "unexpected token %s", t)
		goto end
	}
	next = p.read()
	if next.Type != TokenRightDelim {
		err = p.errorf(t, "unexpected token %s", t)
		goto end
	}
	for {
		read, err = p.readv(t)
		if err != nil {
			goto end
		}
		tokens = append(tokens, read...)
		if len(read) > 1 {
			// Check the token that preceded the matching identifier. For
			// section start and inverse tokens we increase the stack, otherwise
			// decrease.
			tt := read[len(read)-2]
			//goland:noinspection GoSwitchMissingCasesForIotaConsts
			switch tt.Type {
			case TokenSectionStart, TokenSectionInverse:
				stack++
			case TokenSectionEnd:

				stack--
			}
		}
		if stack == 0 {
			break
		}
	}
	nodes, err = subParser(tokens[:len(tokens)-3]).Parse()
	if err != nil {
		goto end
	}
	section = &SectionNode{
		Name:     t.Value,
		Inverted: inverse,
		Elems:    nodes,
	}
end:
	return section, err
}

// parsePartial parses a partial block. It is assumed that the next read should
// return a t_ident token.
func (p *Parser) parsePartial() (node Node, err error) {
	var isDynamic bool
	var path []string

	// At this point we've just seen TokenPartial ("{{>")
	// We need to track if we're in a standalone context
	var isStandalone bool

	t := p.read()
	for {
		switch t.Type {
		case TokenEOF:
			err = p.errorf(t, "unexpected end of template while parsing partial")
			goto end

		case TokenDynamicStart:
			isDynamic = true
			t = p.read()
			switch t.Type {
			case TokenDynamicStart:
				// If we're already in dynamic mode, this is a double asterisk -
				// return an empty partial node
				node = &PartialNode{
					name:      "",
					isDynamic: true,
					//IsStandalone: false,
					//indent:       "",
				}
				goto end
			default:
				continue
			}

		case TokenIdentifier:
			path = append(path, t.Value)

		case TokenDot:
			// Continue to next token

		case TokenRightDelim:
			if isStandalone {
				// If we started standalone, verify by checking what follows
				next := p.peek()
				// Only keep standalone status if followed by newline token
				if next.Type != TokenNewlineText {
					// If next token isn't a newline or pure whitespace,
					// this isn't standalone
					isStandalone = false
				} else {
					// Consume the newline token since we've confirmed standalone
					p.read()
				}
			}
			node = &PartialNode{
				name:      strings.Join(path, "."),
				isDynamic: isDynamic,
				// A partial is standalone if it has only whitespace before it on the line
				// and either a newline or only whitespace follows it
				//IsStandalone: IsStandalone,
			}
			goto end

		default:
			err = p.errorf(t, "unexpected token %s", t)
			goto end
		}
		t = p.read()
	}

end:
	return node, err
}

// NewParser creates a new Parser using the supplied lexer.
func NewParser(l *Lexer) *Parser {
	return &Parser{lexer: l}
}

// subParser creates a new parser with a pre-defined token buffer.
func subParser(b []Token) *Parser {
	return &Parser{buf: append(b, Token{Type: TokenEOF})}
}

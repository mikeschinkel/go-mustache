package mustache

import (
	"fmt"
	"strings"
)

var _ Node = (*TextNode)(nil)

// TextNode represents literal text in a mustache template.
// This is any content outside of mustache tags ({{...}}).
// TextNode is implemented as a string alias, storing the raw text content.
type TextNode string

func (tn TextNode) Clone() Node {
	return tn
}

// Render implements the Node interface for TextNode.
// It writes the text content to the writer, character by character,
// and marks the writer as having non-tag content if non-whitespace is encountered.
// This marking is important for proper standalone tag handling.
//
//goland:noinspection GoUnusedParameter
func (tn TextNode) Render(t *Template, w *Writer, c ...interface{}) error {
	text := string(tn)
	for _, r := range text {
		if !whitespace(r) {
			w.text()
		}
		err := w.WriteRune(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// String returns a string representation of the TextNode for debugging.
func (tn TextNode) String() string {
	return fmt.Sprintf("[text: %q]", string(tn))
}

// Indent returns the common whitespace prefix for all non-empty lines in the TextNode
func (tn TextNode) Indent(lines ...[]string) (s string) {
	if len(lines) != 0 {
		s, _ = tn.indent(lines[0])
		goto end
	}
	s, _ = tn.indent(strings.Split(string(tn), "\n"))
end:
	return s
}

// RemoveIndent removes the common indentation from all lines in the TextNode
func (tn TextNode) RemoveIndent(lines ...[]string) (text TextNode) {
	var ll []string
	if len(lines) != 0 {
		ll = lines[0]
	} else {
		s := string(tn)
		ll = strings.Split(s, "\n")
	}
	indent, empties := tn.indent(ll)
	if indent == "" {
		goto end
	}

	for i, line := range ll {
		if empties[i] {
			continue
		}
		if !strings.HasPrefix(line, indent) {
			continue
		}
		ll[i] = line[len(indent):]
	}
	text = TextNode(strings.Join(ll, "\n"))
end:
	return text
}

func (tn TextNode) indent(lines []string) (indent string, empties []bool) {
	empties = make([]bool, len(lines))
	// Find the shortest indent for all lines
	for i, line := range lines {
		lineIndent := leadingWhitespace(line)
		if len(lineIndent) == len(line) {
			empties[i] = true
			continue
		}
		if indent != "" && len(indent) < len(lineIndent) {
			continue
		}
		indent = lineIndent
	}
	return indent, empties
}

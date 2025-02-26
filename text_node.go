package mustache

import (
	"fmt"
)

// TextNode represents literal text in a mustache template.
// This is any content outside of mustache tags ({{...}}).
// TextNode is implemented as a string alias, storing the raw text content.
type TextNode string

// Render implements the Node interface for TextNode.
// It writes the text content to the writer, character by character,
// and marks the writer as having non-tag content if non-whitespace is encountered.
// This marking is important for proper standalone tag handling.
//
//goland:noinspection GoUnusedParameter
func (n TextNode) Render(t *Template, w *Writer, c ...interface{}) error {
	text := string(n)
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
func (n TextNode) String() string {
	return fmt.Sprintf("[text: %q]", string(n))
}

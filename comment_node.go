package mustache

import (
	"fmt"
)

// CommentNode represents a comment tag in a mustache template, like "{{! This is a comment }}".
// Comments are completely ignored during rendering and produce no output.
// The content of the comment is stored as a string for debugging purposes.
type CommentNode string

// Render implements the Node interface for CommentNode.
// Comments are ignored during rendering, so this simply marks that a tag was
// processed (for proper whitespace handling) and returns nil.
//
//goland:noinspection GoUnusedParame
//goland:noinspection GoUnusedParameter
func (n CommentNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.tag()
	return nil
}

// String returns a string representation of the CommentNode for debugging.
func (n CommentNode) String() string {
	return fmt.Sprintf("[comment: %q]", string(n))
}

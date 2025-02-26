package mustache

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// Writer is a specialized writer for mustache templates that provides
// additional functionality for tracking whether text and tags have been
// written and for managing whitespace according to the Mustache spec.
//
// Writer handles the buffering and flushing of output according to Mustache's
// whitespace rules, particularly for standalone tags.
type Writer struct {
	// HasText indicates whether non-tag text has been written to this writer
	HasText bool
	// HasTag indicates whether a tag has been written to this writer
	HasTag bool
	// w is the underlying io.Writer where output is eventually written
	w io.Writer
	// b is a buffered writer wrapping w
	b *bufio.Writer
	// sb is a string builder for accumulating and inspecting content
	sb *strings.Builder
}

// NewWriter creates a new Writer wrapping the provided io.Writer.
// The returned Writer handles mustache-specific writing needs, including
// tracking whether text and tags have been written and managing whitespace.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		HasText: false,
		HasTag:  false,
		w:       w,
		b:       bufio.NewWriter(w),
		sb:      &strings.Builder{},
	}
}

// text marks this writer as having written non-tag text.
// This method is called by text nodes to indicate their presence.
func (w *Writer) text() {
	w.HasText = true
}

// tag marks this writer as having written a tag.
// This method is called by tag nodes (variables, sections, etc.)
// to indicate their presence.
func (w *Writer) tag() {
	w.HasTag = true
}

// reset returns the writer to its initial state.
// This clears the HasText and HasTag flags and reinitializes the string builder.
func (w *Writer) reset() {
	w.HasTag = false
	w.HasText = false
	w.sb = &strings.Builder{}
}

// Flush writes any buffered data to the underlying io.Writer.
// If a line contains only tags and no text, the line is omitted
// according to the Mustache spec's rules for standalone tags.
// After flushing, the writer's state is reset.
func (w *Writer) Flush() (err error) {
	defer w.reset()
	if w.HasTag && !w.HasText {
		w.b.Reset(w.w)
		return nil
	}
	return w.b.Flush()
}

// WriteRune writes a single Unicode code point to the writer.
// If the rune is a newline, the writer automatically flushes.
// This behavior helps implement the standalone tag rules.
func (w *Writer) WriteRune(r rune) error {
	_, err := w.b.WriteRune(r)
	if err != nil {
		return err
	}
	if r == '\n' {
		return w.Flush()
	}
	return nil
}

// WriteString writes a string to the writer.
// This method converts the string to bytes and delegates to Write.
func (w *Writer) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

// Write writes len(b) bytes from b to the underlying writer.
// The bytes are written rune by rune to properly handle Unicode
// and to ensure proper flushing on newlines.
func (w *Writer) Write(b []byte) (int, error) {
	w.sb.Write(b)
	for i, r := range bytes.Runes(b) {
		err := w.WriteRune(r)
		if err != nil {
			return i, err
		}
	}
	return len(b), nil
}

// String returns the accumulated content as a string.
// This is primarily used for debugging and testing.
func (w *Writer) String() string {
	return w.sb.String()
}

package mustache

import (
	"fmt"
)

var _ Node = (*LeadingWhitespaceNode)(nil)

type LeadingWhitespaceNode string

func (n LeadingWhitespaceNode) Clone() Node {
	return n
}

//goland:noinspection GoUnusedParameter
func (n LeadingWhitespaceNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	// Nothing gets rendered here, it is just a placeholder
	return err
}

func (n LeadingWhitespaceNode) String() string {
	return fmt.Sprintf(`[indent: %q]`, string(n))
}

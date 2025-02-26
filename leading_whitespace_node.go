package mustache

import (
	"fmt"
)

type LeadingWhitespaceNode string

//goland:noinspection GoUnusedParameter
func (n LeadingWhitespaceNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	// Nothing gets rendered here, it is just a placeholder
	return err
}

func (n LeadingWhitespaceNode) String() string {
	return fmt.Sprintf(`[indent: %q]`, string(n))
}

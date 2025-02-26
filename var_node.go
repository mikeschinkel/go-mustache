package mustache

import (
	"fmt"
)

// VarNode represents a variable tag in a mustache template, like "{{name}}".
// When rendered, it looks up the variable in the context and renders its value.
// This handles both escaped "{{name}}" and unescaped "{{{name}}}" or "{{&name}}" tags.
type VarNode struct {
	// Name is the variable name or path to look up (e.g., "user.name")
	Name string
	// Escape determines whether HTML special characters should be escaped
	// (true for "{{name}}", false for "{{{name}}}" or "{{&name}}")
	Escape bool
}

// Render implements the Node interface for VarNode.
// It looks up the variable name in the context and writes its value to the writer.
// If HTML escaping is enabled, it escapes any HTML special characters.
//
// The lookup first tries in the immediate context. If that fails, it tries
// the root context to support absolute references in nested sections.
func (n *VarNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.text()
	v, _, err := t.Lookup(n.Name, c...)
	if err != nil {
		// Handle when inner names are t.root-absolute rather than c-relative, e.g.
		// {{#foo.bar}}{{foo.bar.baz}}{{/foo.bar}}
		v, _, err = t.Lookup(n.Name, t.root)
	}
	if err != nil {
		switch {
		case !t.silentMiss:
			return err
		case !t.injectOnMiss:
			return nil
		default:
			injectError(w.w, err)
		}
	}
	// If the value is present but 'falsy', such as a false bool, or a zero int,
	// we still want to render that value.
	if v != nil {
		if n.Escape {
			v = escape(fmt.Sprintf("%v", v))
		}
		write(w, v)
		return nil
	}
	return fmt.Errorf("failed to lookup %s", n.Name)
}

// String returns a string representation of the VarNode for debugging.
func (n *VarNode) String() string {
	return fmt.Sprintf("[var: %q escaped: %t]", n.Name, n.Escape)
}

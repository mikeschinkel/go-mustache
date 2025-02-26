package mustache

import (
	"fmt"
	"reflect"
)

var _ ChildrenGetter = (*SectionNode)(nil)

// SectionNode represents a section tag in a mustache template, like "{{#section}}...{{/section}}"
// or an inverted section "{{^section}}...{{/section}}".
//
// Sections have multiple behaviors depending on the value they reference:
// - If the value is falsy (false, zero, empty, nil), the section is not rendered
// - If the value is a non-empty array/slice, the section is rendered once for each item
// - If the value is a map or struct, the section is rendered with that as context
// - If the value is truthy, the section is rendered with the current context
//
// Inverted sections ({{^section}}) have the opposite behavior - they render when the
// value is falsy or an empty array/slice.
type SectionNode struct {
	// Name is the section identifier to look up in the context
	Name string
	// Inverted determines if this is a normal section ({{#name}}) or inverted section ({{^name}})
	Inverted bool
	// Elems contains the child nodes inside the section
	Elems []Node
}

// GetChildren returns the section's child nodes.
// This method implements the ChildrenGetter interface.
func (n *SectionNode) GetChildren() []Node {
	return n.Elems
}

// Render implements the Node interface for SectionNode.
// It handles the complex logic of section rendering based on the value type:
// - For arrays/slices: iterates and renders once per item
// - For other types: renders once with the value as context
// - For inverted sections: renders only if the value is falsy
//
// The method also manages lookups in both local and root contexts to support
// both relative and absolute references.
func (n *SectionNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.tag()
	defer w.tag()
	// Helper function to render all child elements with the given context
	elemFn := func(v ...interface{}) error {
		n.Elems = fixWhitespace(n.Elems, nil)
		for _, elem := range n.Elems {
			err := elem.Render(t, w, append(v, c...)...)
			if err != nil {
				if !t.silentMiss {
					return fmt.Errorf("failed to render element %s; %w", n.Name, err)
				} else if !t.injectOnMiss {
					return nil
				} else {
					injectError(w.w, err)
				}
			}
		}
		return nil
	}

	// Try to lookup the section name in the current context
	v, ok, err := t.Lookup(n.Name, c...)
	if err != nil {
		// If that fails, try the root context (for absolute references)
		v, ok, err = t.Lookup(n.Name, t.root)
	}
	if err != nil {
		return err
	}

	// If the section should be rendered (normal section with value, or inverted section without)
	if ok != n.Inverted {
		r := reflect.ValueOf(v)
		switch r.Kind() {
		case reflect.Slice, reflect.Array:
			if r.Len() == 0 {
				// Empty slices/arrays are treated like falsy values
				err := elemFn(v)
				if err != nil && !t.silentMiss {
					return err
				}
				return nil
			}
			// For non-empty slices/arrays, iterate and render for each item
			for i := 0; i < r.Len(); i++ {
				err := elemFn(r.Index(i).Interface())
				if err != nil && !t.silentMiss {
					return err
				}
			}
		default:
			// For other truthy values, render once with the value as context
			err := elemFn(v)
			if err != nil && !t.silentMiss {
				return err
			}
		}
		return nil
	}
	return nil
}

// String returns a string representation of the SectionNode for debugging.
func (n *SectionNode) String() string {
	return fmt.Sprintf("[section: %q inv: %t Elems: %s]", n.Name, n.Inverted, n.Elems)
}

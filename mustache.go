// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Kalyvitis

package mustache

import (
	"io"
)

// Parse creates a new template and parses it from the given reader in one step.
// This is a convenience function equivalent to:
//
//	t := New()
//	err := t.Parse(r)
//
// Example:
//
//	file, _ := os.Open("template.mustache")
//	tmpl, err := mustache.Parse(file)
//	if err != nil {
//	    // handle error
//	}
func Parse(r io.Reader) (*Template, error) {
	t := New()
	err := t.Parse(r)
	return t, err
}

// Render parses a template from a reader and immediately renders it with the given context.
// This is a convenience function equivalent to:
//
//	t, err := Parse(r)
//	if err != nil {
//	    return err
//	}
//	return t.Render(w, context...)
//
// This is useful for one-off template rendering where you don't need to
// reuse the template.
//
//goland:noinspection GoUnusedExportedFunction
func Render(r io.Reader, w io.Writer, context ...interface{}) error {
	t, err := Parse(r)
	if err != nil {
		return err
	}
	return t.Render(w, context...)
}

package mustache

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"strings"
)

var _ ChildrenGetter = (*Template)(nil)

// Template represents a parsed mustache template and its components.
// A Template maintains a collection of nodes (Elems) that form the parsed template,
// as well as configuration options like delimiters and partial templates.
// Templates can be created using New() and configured with Option().
type Template struct {
	// Name is the template's identifier, particularly important for partial templates
	Name string
	// Elems contains the parsed nodes that make up the template
	Elems []Node
	// structTag is the tag name used for struct field lookup
	structTag string
	// root stores a reference to the root context object
	root interface{}
	// partials is a map of registered partial templates
	partials map[string]*Template
	// startDelim is the opening delimiter (default "{{")
	startDelim string
	// endDelim is the closing delimiter (default "}}")
	endDelim string
	// silentMiss controls whether missing variables cause errors
	silentMiss bool
	// silentMiss controls whether missing variables cause errors
	logger *slog.Logger
	// injectOnMiss controls whether to inject error messages
	injectOnMiss bool
	// tagIndexCache caches tag lookups for performance
	tagIndexCache map[string]int
	// preprocessors are applied before rendering
	preprocessors []Preprocessor

	overrides *ClonableStack[Overrides]
}

// Clone created a copy of a Template
func (t *Template) Clone() (newTmpl *Template) {
	newTmpl = new(Template)
	*newTmpl = *t

	// Deep copy the node tree
	newTmpl.Elems = make([]Node, len(t.Elems))
	for i, elem := range t.Elems {
		newTmpl.Elems[i] = elem.Clone()
	}

	return newTmpl
}

// GetChildren returns the template's nodes as a slice.
// This method implements the ChildrenGetter interface.
func (t *Template) GetChildren() []Node {
	return t.Elems
}

// Preprocess applies the given preprocessor to the provided elements.
// This is a convenience method that delegates to the global Preprocess function.
func (t *Template) Preprocess(p Preprocessor, elems []Node) (err error) {
	return Preprocess(p, t, elems)
}

// render applies all preprocessors and then renders the template
func (t *Template) render(w *Writer, context ...interface{}) (err error) {
	errs := make([]error, 0)

	// Apply all registered preprocessors
	for _, p := range t.preprocessors {
		err = Preprocess(p, t, t.Elems)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Continue with normal rendering
	for _, elem := range t.Elems {
		err = elem.Render(t, w, context...)
		if err != nil {
			errs = append(errs, err)
		}
	}

	err = w.Flush()
	if err != nil {
		errs = append(errs, err)
	}

	return t.triageError(w, errors.Join(errs...))
}

// New creates and returns a new Template instance with default settings.
// The returned template uses "{{" and "}}" as delimiters, silently ignores
// missing variables, and includes the SpecConformance preprocessor by default.
//
// Optional configuration can be provided through variadic Option parameters,
// which will be immediately applied to the template.
//
// Example:
//
//	template := mustache.New(
//	    mustache.Name("my-template"),
//	    mustache.Delimiters("<<", ">>"))
func New(options ...Option) *Template {
	t := &Template{
		Elems:         make([]Node, 0),
		partials:      make(map[string]*Template),
		startDelim:    "{{",
		endDelim:      "}}",
		silentMiss:    true,
		tagIndexCache: make(map[string]int),
		preprocessors: make([]Preprocessor, 0),
		overrides:     &ClonableStack[Overrides]{},
	}
	// Ensure Mustang Spec Conformance across nodes and partials
	t.AddPreprocessor(&SpecConformance{})

	t.Option(options...)
	return t
}

// Option applies options to the current template t.
func (t *Template) Option(options ...Option) {
	for _, optionFn := range options {
		optionFn(t)
	}
}

// Parse processes the content from the provided reader and builds
// a parse tree that represents the template. The template uses its configured
// delimiters to identify mustache tags in the input.
//
// This is the core parsing method that other parsing helpers (ParseString,
// ParseBytes) delegate to.
//
// Returns an error if reading from the reader fails or if the template
// contains syntax errors.
func (t *Template) Parse(r io.Reader) (err error) {
	var l *Lexer
	var p *Parser
	var elems []Node

	b, err := io.ReadAll(r)
	if err != nil {
		goto end
	}
	l = NewLexer(string(b), t.startDelim, t.endDelim)
	p = NewParser(l)
	p.silentMiss = t.silentMiss
	elems, err = p.Parse()
	if err != nil {
		goto end
	}
	t.Elems = elems
end:
	return err
}

// ParseString parses the provided string into a template.
// This is a convenient wrapper around Parse that converts the string
// to an io.Reader.
//
// Example:
//
//	template := mustache.New()
//	err := template.ParseString("Hello, {{name}}!")
func (t *Template) ParseString(s string) (err error) {
	return t.Parse(strings.NewReader(s))
}

// ParseBytes parses the provided byte slice into a template.
// This is a convenient wrapper around Parse that converts the byte slice
// to an io.Reader.
func (t *Template) ParseBytes(b []byte) error {
	return t.Parse(bytes.NewReader(b))
}

// Render evaluates the template with the given context and writes the result to w.
// The context can be any value (struct, map, etc.) that the template can look up
// values from. The first context object is also stored as the root context and
// is available throughout rendering, including from within sections.
//
// Example:
//
//	data := map[string]interface{}{
//	    "name": "World",
//	    "items": []string{"one", "two", "three"},
//	}
//	template.Render(os.Stdout, data)
func (t *Template) Render(w io.Writer, context ...interface{}) error {
	t.root = context[0]
	return t.render(NewWriter(w), context...)
}

// RenderString evaluates the template with the given context and returns the result as a string.
// This is a convenience wrapper around Render that captures the output in a buffer.
//
// Example:
//
//	result, err := template.RenderString(data)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Println(result)
func (t *Template) RenderString(context ...interface{}) (string, error) {
	b := &bytes.Buffer{}
	err := t.Render(b, context...)
	return b.String(), err
}

// RenderBytes evaluates the template with the given context and returns the result as a byte slice.
// This is a convenience wrapper around Render that captures the output in a buffer.
func (t *Template) RenderBytes(context ...interface{}) (bb []byte, err error) {
	b := &bytes.Buffer{}
	err = t.Render(b, context...)
	if err != nil {
		goto end
	}
	bb = b.Bytes()
end:
	return bb, err
}

// AddPreprocessor registers a preprocessor to be applied before rendering.
// Preprocessors can modify the template's parse tree before it's rendered.
// The SpecConformance preprocessor is added by default in New().
func (t *Template) AddPreprocessor(p Preprocessor) {
	t.preprocessors = append(t.preprocessors, p)
}

func (t *Template) triageError(w io.Writer, err error) error {
	if err == nil {
		goto end
	}
	if t.logger != nil {
		t.logger.Error(err.Error())
		err = nil
	}
	if t.injectOnMiss && !t.silentMiss {
		injectError(w, err)
		err = nil
	}
end:
	return err
}

func (t *Template) Logger() *slog.Logger {
	return t.logger
}

func (t *Template) getPartial(name string) (partial *Template, found bool, err error) {
	var ok bool
	partial, ok = t.partials[name]
	if ok {
		found = true
		goto end
	}
	if !t.silentMiss {
		err = fmt.Errorf("template '%s' not found", name)
		goto end
	}
end:
	return partial, found, err
}

func (t *Template) pushOverrides(overrides Overrides) (oo Overrides) {
	top := t.overrides.Top().Clone()
	oo = overrides.Clone()
	if top != nil {
		maps.Insert(oo, maps.All(top))
	}
	t.overrides.Push(oo)
	return oo
}

package mustache

import (
	"log/slog"
)

// Option is a function that configures a Template.
// This package uses the functional options pattern for configuring templates.
// Options can be provided when creating a new template or applied later using
// the Template.Option method.
//
// Example:
//
//	// Create with options
//	t := mustache.New(
//	    mustache.Name("my-template"),
//	    mustache.Delimiters("<<", ">>"))
//
//	// Apply options later
//	t.Option(mustache.SilentMiss(false))
//
// For more information on the functional options pattern, see
// Dave Cheney's talk: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(*Template)

// Name sets the name of the template.
// The name is particularly important for partial templates, as it is used
// to reference the template from {{>name}} tags in parent templates.
//
// Example:
//
//	header := mustache.New(mustache.Name("header"))
//	header.ParseString("<h1>{{title}}</h1>")
//
//	// Later, in the parent template: {{>header}}
func Name(n string) Option {
	return func(t *Template) {
		t.Name = n
	}
}

// Delimiters sets the start and end delimiters of the template.
// By default, mustache templates use "{{" and "}}" as delimiters,
// but these can be changed to avoid conflicts with other templating
// systems or special content.
//
// Example:
//
//	// Use "<<" and ">>" as delimiters
//	t := mustache.New(mustache.Delimiters("<<", ">>"))
//	t.ParseString("Hello, <<name>>!")
func Delimiters(start, end string) Option {
	return func(t *Template) {
		t.startDelim = start
		t.endDelim = end
	}
}

// Partial registers a template as a partial that can be referenced
// from within this template using the {{>name}} syntax.
//
// It is important that the partial template has been given a name using
// the Name option, as this name is used to look up the partial.
//
// Example:
//
//	header := mustache.New(mustache.Name("header"))
//	header.ParseString("<h1>{{title}}</h1>")
//
//	footer := mustache.New(mustache.Name("footer"))
//	footer.ParseString("<footer>{{copyright}}</footer>")
//
//	main := mustache.New(
//	    mustache.Partial(header),
//	    mustache.Partial(footer))
//	main.ParseString("{{>header}}Content{{>footer}}")
func Partial(p *Template) Option {
	return func(t *Template) {
		t.partials[p.Name] = p
	}
}

// Errors enables missing variable errors. This option is deprecated.
// Please use SilentMiss(false) instead for equivalent behavior.
//
//goland:noinspection GoUnusedExportedFunction
func Errors() Option {
	return func(t *Template) {
		t.silentMiss = false
	}
}

// SilentMiss controls whether missing variable lookups cause errors during rendering.
// By default, templates use SilentMiss(true), which means missing variables are
// simply rendered as empty strings.
//
// When set to false, a missing variable will stop rendering and return an error.
// This is useful during development to catch template/data mismatches.
//
// Example:
//
//	// Report errors for missing variables
//	t := mustache.New(mustache.SilentMiss(false))
//	_, err := t.RenderString(data)
//	if err != nil {
//	    log.Fatalf("Template rendering failed: %v", err)
//	}
func SilentMiss(silent bool) Option {
	return func(t *Template) {
		t.silentMiss = silent
	}
}

// InjectOnMiss controls whether error messages for missing variables are injected
// into the output. This only has an effect when SilentMiss is true (the default).
//
// This is primarily useful for debugging templates, as it will show exactly
// where variables are missing directly in the rendered output.
//
//goland:noinspection GoUnusedExportedFunction
func InjectOnMiss(inject bool) Option {
	return func(t *Template) {
		t.injectOnMiss = inject
	}
}

func Logger(logger *slog.Logger) Option {
	return func(t *Template) {
		t.logger = logger
	}
}

// StructTag sets the struct tag name used when looking up fields in struct values.
// By default, the package uses the tag name "mustache".
//
// Example with custom tag:
//
//	// Struct with custom tag
//	type User struct {
//	    FirstName string `json:"first_name"`
//	    LastName  string `json:"last_name"`
//	}
//
//	// Use the JSON tag instead of the default "mustache" tag
//	t := mustache.New(mustache.StructTag("json"))
//	t.ParseString("Hello, {{first_name}} {{last_name}}!")
//
//goland:noinspection GoUnusedExportedFunction
func StructTag(tag string) Option {
	return func(t *Template) {
		t.structTag = tag
	}
}

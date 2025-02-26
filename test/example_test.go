// Copyright (c) 2014 Alex Kalyvitis

package mustache_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/alexkappa/mustache"
)

func ExampleTemplate_basic() {
	var err error
	var output string

	context := map[string]interface{}{
		"foo": true,
		"bar": "bazinga!",
	}

	template := mustache.New()
	err = template.ParseString(`{{#foo}}{{bar}}{{/foo}}`)
	if err != nil {
		goto end
	}

	output, _ = template.RenderString(context)
	fmt.Println(output)
	// Output: bazinga!

end:
	if err != nil {
		panic(err)
	}
}

func ExampleTemplate_partials() {
	var err error
	var template *mustache.Template

	context := map[string]interface{}{
		"foo": true,
		"bar": "bazinga!",
	}

	partial := mustache.New(Name("partial"))
	err = partial.ParseString(`{{bar}}`)
	if err != nil {
		goto end
	}

	template = mustache.New(Partial(partial))
	err = template.ParseString(`{{#foo}}{{>partial}}{{/foo}}`)
	if err != nil {
		goto end
	}

	err = template.Render(os.Stdout, context)
	// Output: bazinga!

end:
	if err != nil {
		log.Fatal(err)
	}

}

func ExampleTemplate_reader() {
	var t *mustache.Template

	f, err := os.Open("template.mustache")
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "failed to open file: %s\n", err)
		goto end
	}

	t, err = mustache.Parse(f)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "failed to parse template: %s\n", err)
		goto end
	}
	err = t.Render(os.Stdout, nil)

end:
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleTemplate_http() {
	var err error

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://example.com?foo=bar&bar=one&bar=two", nil)

	template := mustache.New()
	err = template.ParseString(`
<ul>{{#foo}}<li>{{.}}</li>{{/foo}}</ul>
<ul>{{#bar}}<li>{{.}}</li>{{/bar}}</ul>`)
	if err != nil {
		goto end
	}

	err = template.Render(writer, request.URL.Query())
	if err != nil {
		goto end
	}

	fmt.Println(writer.Body.String())
	// Output:
	// <ul><li>bar</li></ul>
	// <ul><li>one</li><li>two</li></ul>
end:
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleOption() {
	var err error
	var body, template *mustache.Template

	context := map[string]interface{}{
		"title":   "Mustache",
		"content": "Logic less templates with Mustache!",
	}

	title := mustache.New(Name("header")) // instantiate and name the template
	err = title.ParseString("{{title}}")  // parse a template string
	if err != nil {
		goto end
	}

	body = mustache.New()
	body.Option(Name("body")) // options can be defined after we instantiate too
	err = body.ParseString("{{content}}")
	if err != nil {
		goto end
	}
	template = mustache.New(
		mustache.Delimiters("|", "|"), // set the mustache delimiters to | instead of {{
		mustache.SilentMiss(false),    // return an error if a variable lookup fails
		Partial(title),                // register a partial
		Partial(body))                 // and another one...

	err = template.ParseString("|>header|\n|>body|")
	if err != nil {
		goto end
	}

	err = template.Render(os.Stdout, context)
	if err != nil {
		goto end
	}
	// Output: Mustache
	// Logic less templates with Mustache!
end:
}

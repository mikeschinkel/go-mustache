# Mustache Template Classifier

A package for the Go programming language to classify the lines and line segments of _(almost?)_ any valid [Mustache](https://mustache.github.io/) template.

## Overview

The Mustache Template Classifier is a specialized package for analyzing and classifying Mustache templates. It preprocesses templates and assigns line types and segments to each line, which can be used to help understand what constitutes proper indentation, whitespace handling, and newline processing according to the [Mustache specification](https://github.com/mustache/spec).

This package was created to handle the more complex aspects of Mustache rendering, especially when dealing with:

- Standalone tags (tags on their own line)
- Indentation preservation
- Whitespace management
- Newline elision

## Features

- Line classification (empty, whitespace, text, inline, standalone)
- Segment-based token analysis
- Proper handling of all Mustache tag types
  - Sections (`{{#section}}`)
  - Inverted sections (`{{^section}}`)
  - Comments (`{{! comment }}`) 
  - Partials (`{{>partial}}`)
  - Variables (`{{variable}}`)
  - Unescaped variables (`{{{variable}}}` and `{{&variable}}`)
  - Delimiters (`{{=< >=}}`)
  - Blocks (`{{$block}}`)
  - Parents (`{{<parent}}`)
- Support for triple braces
- Support for custom delimiters
- Enclosing tag matching

## Installation

```bash
go get github.com/alexkappa/mustache/classify
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/alexkappa/mustache/classify"
)

func main() {
	// Create a new template classifier
	template := "{{#section}}\n  Content\n{{/section}}"
	tc := classify.NewTemplateClassifier(template)

	// Classify the template
	lines, err := tc.Classify()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// Process the classified lines
	for i, line := range lines {
		fmt.Printf("Line %d: Type=%s, Segments=%+v\n", i, line.Type, line.Segments)
	}
}
```

## Line Types

The classifier categorizes each line into one of the following types:

- `EmptyLine`: A line with no content
- `WhitespaceLine`: A line containing only whitespace
- `TextLine`: A line containing only text content
- `InlineLine`: A line with mixed content (text and tags)
- `StandaloneLine`: A line with a tag that stands alone (possibly with leading/trailing whitespace)

## Segment Types

Each line is further broken down into segments, which can be:

- `TextContent`: Regular template text
- `Whitespace`: Spaces, tabs, etc.
- `CompleteTag`: A standalone tag
- `BeginTag`: An opening tag (e.g., `{{#section}}`)
- `EndTag`: A closing tag (e.g., `{{/section}}`)
- `MultilineBegin`: First line of a multiline tag
- `MultilineMiddle`: Middle line of a multiline tag
- `MultilineEnd`: Last line of a multiline tag

## Why? / For What Use-Case?

This module's parent contains a fork of a Go package — [github.com/alexkappa/mustache](https://github.com/alexkappa/mustache). This fork is intended to add support for all the Mustache conformance tests found [here](https://github.com/mustache/spec/tree/master/specs).

This repo's original project — from [alexkappa/mustache](https://github.com/alexkappa/mustache) uses a lexer and then a parser to process mustache template files and then a renderer. As I worked on my fork I added a preprocessor between the parser and renderer to handle some of the more arcane rules for indentation and eliding newlines. However, at one point I into a problem reasoning about the parser, especially given how the lexer and parser interact making it hard to follow.

Given I found myself stalled on being able to handle all the conformance tests because my inability to reason about the arcane rules for indentation and eliding newlines I decided to create a small sub-project to **classify all lines in a template** with a line type and with a slice of token slices — which I called `Segments` — so that I could better reason about them. That is the code that exists in this module.

My plan was to use this in my preprocessor for the parent module when I needed to decide on indentation and/or eliding newlines but in retrospect using two (2) parsers to parse the same template for one rendering is probably not a great idea. At the time of this writing I think I should just revisit fixing alexkappa's parser now that I better understand Mustache's requirements.

**_Still_**, since I did all the work to get this template classifier to pass the over 100 tests comprising various scenarios I felt it would be a shame to just throw it all away. 

So **that** is the _why_ for this repo. I do not know if anyone will ever find it useful, but if someone does then I guess it will have been worth me not throwing it away. 🤷‍♂️

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the same license as github.com/alexkappa/mustache.

## Acknowledgments

This package was developed as an extension to [alexkappa/mustache](https://github.com/alexkappa/mustache) to provide better support for the full Mustache specification.
# Mustache Template Classifier

A template classification library for the [mustache templating language](https://mustache.github.io/) in Go.

## Overview

The Mustache Template Classifier is a specialized package for analyzing and classifying Mustache templates. It preprocesses templates and assigns line types and segments to each line, which helps with proper indentation, whitespace handling, and newline processing according to the [Mustache specification](https://github.com/mustache/spec).

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the same license as github.com/alexkappa/mustache.

## Acknowledgments

This package was developed as an extension to [alexkappa/mustache](https://github.com/alexkappa/mustache) to provide better support for the full Mustache specification.
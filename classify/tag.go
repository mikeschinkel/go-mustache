package classify

import (
	"fmt"
)

// Tag represents a Mustache tag with its identifier and type. This is used to
// along with a stack in TemplateClassifier to track enclosing tags (like
// sections, blocks, etc.) when parsing templates.
type Tag struct {
	// Identifier is the name of the tag (e.g., "section" in {{#section}}).
	Identifier string

	// Type is the semantic type of the tag (e.g., SectionTag, BlockTag).
	Type TagType
}

// String returns a human-readable representation of Tag for error messages and
// test output.
func (t Tag) String() string {
	return fmt.Sprintf("[tag: {name: '%s', type: '%s'}]", t.Identifier, t.Type)
}

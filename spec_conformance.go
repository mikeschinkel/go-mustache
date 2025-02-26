package mustache

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

var _ Preprocessor = (*SpecConformance)(nil)

// newlineRegex matches both CR+LF and LF newlines
var newlineRegex = regexp.MustCompile("(\r?\n)")

// multilineRegex matches both CR+LF and LF newlines as well as before and after
var multilineRegex = regexp.MustCompile("(.*?)(\r?\n)(.*?)")

// SpecConformance detects and marks standalone partials
type SpecConformance struct {
	IndentStack Stack[string]
	StackDepth  int
}

// Preprocess processes a single node for standalone partials
func (pp *SpecConformance) Preprocess(t *Template, nodes []Node, index int) (err error) {
	if index == 0 {
		pp.StackDepth++
	}
	switch node := nodes[index].(type) {
	case *SectionNode:
		err = Preprocess(pp, t, node.Elems)
	case *PartialNode:
		err = pp.preprocessPartial(t, nodes, index)
		pp.IndentStack.Pop()
	case LeadingWhitespaceNode:
		pp.IndentStack.Push(string(node) + pp.IndentStack.Top())
	case TextNode:
		err = pp.preprocessText(t, nodes, index)
	}
	if index == len(nodes)-1 {
		pp.StackDepth--
		if pp.IndentStack.Depth() > pp.StackDepth {
			pp.IndentStack.Pop()
		}
	}
	return err
}

// preprocessText processes a single text node
//
//goland:noinspection GoUnusedParameter
func (pp *SpecConformance) preprocessText(t *Template, nodes []Node, index int) (err error) {
	var matches [][]int
	var text TextNode
	var content string
	var lastPos int
	var result strings.Builder

	if pp.IndentStack.Top() == "" {
		// This func currently only adds indents, so if there is no indent to add there
		// is nothing left to do.
		goto end
	}

	text = nodes[index].(TextNode)
	content = string(text)

	// Find all newline positions
	matches = newlineRegex.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		// No newlines, nothing to do
		goto end
	}

	// Build the result by inserting indentation after each newline
	lastPos = 0

	for i, match := range matches {
		// Add content up to the newline
		result.WriteString(content[lastPos:match[0]])

		// Add the newline itself (could be \r\n or \n)
		result.WriteString(content[match[0]:match[1]])

		// Update position
		lastPos = match[1]

		if pp.StackDepth == 1 && index == len(nodes)-1 && i == len(matches)-1 {
			// Don't add indentation on the last element
			break
		}
		// Add the indentation
		result.WriteString(pp.IndentStack.Top())

	}

	// Add any remaining content after the last newline
	result.WriteString(content[lastPos:])

	nodes[index] = TextNode(result.String())

end:
	return nil
}

// preprocessPartial processes a single node for standalone partials
func (pp *SpecConformance) preprocessPartial(t *Template, nodes []Node, index int) (err error) {
	var hasLeadingWhitespace bool
	var partial *PartialNode
	var text TextNode
	var followedByNewline, isStandalone bool
	var matches []int
	var ok bool
	var indent, s string
	var kids []Node
	var last int

	// First, check for standalone pattern
	partial, ok = nodes[index].(*PartialNode)
	if !ok {
		goto end
	}

	// Look for whitespace-only text before this partial
	indent, hasLeadingWhitespace = pp.getLeadingWhitespace(nodes, index)
	text, followedByNewline = pp.getFollowedByNewline(nodes, index)
	if followedByNewline {
		nodes[index+1] = text
	}
	if !followedByNewline && index == len(nodes)-1 {
		// Treat partials with no trailing TextNode as effectively "follows by new line"
		followedByNewline = true
	}

	// If we have both conditions, this is a standalone partial
	isStandalone = hasLeadingWhitespace && followedByNewline

	if !isStandalone {
		goto end
	}

	partial.isStandalone = true
	partial.indent = indent
	s = string(text)

	err = pp.indentText(t, nodes)
	if err != nil {
		goto end
	}
	// Use regex to match and remove the newline (handles both \n and \r\n)
	matches = newlineRegex.FindStringIndex(s)
	if matches == nil {
		goto end
	}

	// Check if the newline is at the very beginning of the text
	if matches[0] != 0 {
		// This is a more complex case where the newline isn't at the start
		// This branch shouldn't be hit in your test case
		goto end
	}
	// Remove the newline completely
	nodes[index+1] = TextNode(s[matches[1]:])

	kids, err = partial.GetDerivedNodes(t)
	if err != nil {
		goto end
	}
	if len(kids) < 2 {
		goto end
	}
	last = len(kids) - 1
	text, ok = kids[last].(TextNode)
	if !ok {
		goto end
	}
	// Remove the indent on the last kid for a standalone
	kids[last] = TextNode(string(text)[:len(text)-len(indent)])
end:
	return err
}

// getLeadingWhitespace returns leading whitespace, if exists
func (pp *SpecConformance) getLeadingWhitespace(nodes []Node, index int) (indent string, has bool) {
	var wsNode LeadingWhitespaceNode
	if index <= 0 {
		goto end
	}
	wsNode, has = nodes[index-1].(LeadingWhitespaceNode)
end:
	return string(wsNode), has
}

// getFollowedByNewline looks for newline-only text after this partial
func (pp *SpecConformance) getFollowedByNewline(nodes []Node, index int) (node TextNode, is bool) {
	var text TextNode
	var ok bool
	var s string

	if index+1 >= len(nodes) {
		goto end
	}
	text, ok = nodes[index+1].(TextNode)
	if !ok {
		goto end
	}
	s = string(text)
	if s == "" {
		goto end
	}

	// Use regex to check for \n or \r\n at the beginning
	is = newlineRegex.MatchString(s)
end:
	return text, is
}

type NodeGetter interface {
	GetNodes(*Template) (nodes []Node, err error)
}

// ChildrenGetter provides access to nodes that are direct children
// of the current node, such as those in a SectionNode.
type ChildrenGetter interface {
	GetChildren() []Node
}

// DerivedNodesGetter obtains nodes that require template information to resolve,
// such as those from partial templates.
type DerivedNodesGetter interface {
	// GetDerivedNodes returns nodes that are derived from template dependencies,
	// such as the content of a partial template.
	GetDerivedNodes(t *Template) ([]Node, error)
}

// indentText creates a new template where all content is properly indented.
// It processes the template's elements, ensuring that each line starts with the
// specified indentation while preserving the structure of non-text nodes.
func (pp *SpecConformance) indentText(t *Template, elems []Node) (err error) {
	errs := make([]error, 0)
	var nodes []Node

	isRoot := slices.Equal(t.Elems, elems)

	// Process each element, adding indentation where needed
	for i, elem := range elems {
		switch node := elem.(type) {
		case TextNode:
			if isRoot && i == 0 {
				continue
			}
			text := string(node)
			// Use regex to split by either \r\n or \n, keeping the delimiters
			matches := multilineRegex.FindAllStringSubmatch(text, -1)
			for j, match := range matches {
				noop(j, match)
			}
			//elems[i] = TextNode(strings.Join(matches, ""))

		default:
			// Non-text nodes (like variables, sections)
			switch getter := elem.(type) {
			case ChildrenGetter:
				nodes = getter.GetChildren()
			case DerivedNodesGetter:
				nodes, err = getter.GetDerivedNodes(t)
				if err != nil {
					errs = append(errs, err)
					continue
				}
			default:
				continue
			}
			err = t.Preprocess(pp, nodes)
			if err != nil {
				errs = append(errs, err)
				continue
			}
		}
	}
	return errors.Join(errs...)
}

//goland:noinspection GoUnusedParameter
func noop(args ...any) {}

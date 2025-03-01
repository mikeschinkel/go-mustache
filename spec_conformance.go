package mustache

import (
	"regexp"
	"slices"
	"strings"
)

var _ Preprocessor = (*SpecConformance)(nil)

// newlineRegex matches both CR+LF and LF newlines
var newlineRegex = regexp.MustCompile("(\r?\n)")

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
		// Keep existing partial handling
		err = pp.preprocessStandalone(t, nodes, index, node)
		pp.IndentStack.Pop()
	case *OverrideNode:
		err = pp.preprocessOverride(t, node)
		pp.IndentStack.Pop()
	case *InheritNode:
		// Handle InheritNode as standalone tag
		err = pp.preprocessStandalone(t, nodes, index, node)
		pp.IndentStack.Pop()
	case LeadingWhitespaceNode:
		pp.IndentStack.Push(string(node) + pp.IndentStack.Top())
	case TextNode:
		err = pp.preprocessText(t, nodes, index, node)
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
func (pp *SpecConformance) preprocessText(t *Template, nodes []Node, index int, text TextNode) (err error) {
	var matches [][]int
	var content string
	var lastPos int
	var result strings.Builder

	if pp.IndentStack.Top() == "" {
		// This func currently only adds indents, so if there is no indent to add there
		// is nothing left to do.
		goto end
	}

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

		//if pp.StackDepth == 1 && index == len(nodes)-1 && i == len(matches)-1 {
		noop(i)
		if pp.StackDepth == 1 && index == len(nodes)-1 {
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
	return err
}

// getLeadingWhitespace returns leading whitespace, if exists
func (pp *SpecConformance) getLeadingWhitespace(nodes []Node, index int) (indent string, has bool) {
	var wsNode LeadingWhitespaceNode
	switch index {
	case 0:
		// No prior TextNode means effectively that it "has leading whitespace."
		has = true
	default:
		wsNode, has = nodes[index-1].(LeadingWhitespaceNode)
	}
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

// indentNodes creates a new template where all content is properly indented.
// It processes the template's elements, ensuring that each line starts with the
// specified indentation while preserving the structure of non-text nodes.
func (pp *SpecConformance) indentNodes(t *Template, nodes []Node) (err error) {
	errs := NewMultiErr()
	isRoot := slices.Equal(t.Elems, nodes)

	// Process each element, adding indentation where needed
	for i, node := range nodes {
		_, ok := node.(TextNode)
		if ok && isRoot && i == 0 {
			continue
		}
		err = pp.indentNode(t, nodes, i, node)
		if err != nil {
			errs.Add(err)
		}
	}
	return errs.Err()
}

func (pp *SpecConformance) indentNode(t *Template, nodes []Node, index int, elem Node) (err error) {

	switch node := elem.(type) {
	case TextNode:
		err = pp.preprocessText(t, nodes, index, node)

	default:
		// Non-text nodes (like variables, sections)
		switch getter := elem.(type) {
		case ChildrenGetter:
			nodes = getter.GetChildren()
		case DerivedNodesGetter:
			nodes, err = getter.GetDerivedNodes(t)
			if err != nil {
				goto end
			}
		default:
			goto end
		}
		err = t.Preprocess(pp, nodes)
	}
end:
	return err
}

func (pp *SpecConformance) preprocessOverride(t *Template, node *OverrideNode) (err error) {
	return pp.indentNodes(t, node.Elems)
}

//goland:noinspection GoUnusedParameter
func noop(args ...any) {}

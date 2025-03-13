package mustache

import (
	"fmt"
	"strings"
)

var _ StandaloneNode = (*ParentNode)(nil) // Type assertion to verify interface compliance
var _ Node = (*ParentNode)(nil)

// ParentNode represents template inheritance in Mustache.
// It includes a parent template ({{<template}}) and allows for
// block overrides within the parent template.
type ParentNode struct {
	// Name is the parent template
	name string
	// Overrides contains any block overrides defined in this template
	Overrides Overrides
	// isStandalone indicates if this is a standalone tag (only non-whitespace on its line)
	isStandalone bool
	// indent is the whitespace preceding this node when it's a standalone tag
	indent string
}

func NewParentNode(name string, overrides Overrides) *ParentNode {
	return &ParentNode{
		name:      name,
		Overrides: overrides,
	}
}

func (n *ParentNode) Clone() Node {
	return &ParentNode{
		name:         n.name,
		Overrides:    n.Overrides.Clone(),
		isStandalone: n.isStandalone,
		indent:       n.indent,
	}
}

// Render implements the Node interface for ParentNode.
// It loads the parent template, processes any block overrides,
// and renders the combined result.
func (n *ParentNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	var tmpl, parent *Template
	var found bool
	var errs MultiErr
	var overrides Overrides

	// Push our overrides to the stack BEFORE rendering
	// This way they'll be available to all nested ParentNodes
	overrides = t.pushOverrides(n.Overrides)

	// Get the partial template
	tmpl, found, err = t.getPartial(n.name)
	if !found {
		// partial not found
		goto end
	}

	// Clone the template to avoid modifying the original
	// TODO: Does this need to be cloned?
	//goland:noinspection GoDfaErrorMayBeNotNil
	parent = tmpl.Clone()

	// Important: We need to process ANY inheritance within this template FIRST
	// before applying our overrides
	// This ensures that deeper templates are fully resolved before applying overrides

	errs = NewMultiErr()

	// Apply the current overrides to the template
	if len(overrides) > 0 {
		// We need to find and replace block nodes in the parent template
		err = n.applyOverrides(parent, parent.Elems, overrides)
		if err != nil {
			errs.Add(err)
		}
	}

	// Render the parent template with the current context
	// IMPORTANT: Render each element separately to maintain correct order
	//            Using the same Writer ensures content appears in the correct order
	for _, elem := range parent.Elems {
		err = elem.Render(t, w, c...)
		if err != nil {
			errs.Add(err)
		}
	}
	err = errs.Err()

	// Pop overrides after rendering
	t.overrides.Pop()
end:
	return err
}

// String returns a string representation of the ParentNode.
func (n *ParentNode) String() string {
	return fmt.Sprintf("[parent: %q Overrides: %v]", n.name, n.Overrides)
}

// GetDerivedNodes returns the nodes from the template being parent
// Implement DerivedNodesGetter interface to support preprocessing
func (n *ParentNode) GetDerivedNodes(t *Template) (elems []Node, err error) {
	var tmpl, parent *Template
	var found bool

	// Get the partial template
	tmpl, found, err = t.getPartial(n.name)
	if err != nil {
		// partial not found and silentMiss==false
		goto end
	}
	if !found {
		// partial not found and silentMiss==true
		elems = []Node{}
		goto end
	}

	// Clone the template to avoid modifying the original
	// TODO Does this need to be cloned?
	parent = tmpl.Clone()

	// Apply any block overrides
	if len(n.Overrides) > 0 {
		// We need to find and replace block nodes in the parent template
		err = n.applyOverrides(parent, parent.Elems, n.Overrides)
		if err != nil {
			goto end
		}
	}
	elems = parent.Elems
end:
	return elems, err
}

func (n *ParentNode) applyOverride(_ *Template, nodes []Node, index int, overrides Overrides) (err error) {
	var indent string
	var override []Node

	block, ok := nodes[index].(*BlockNode)
	if !ok {
		goto end
	}

	// First, check if this is a block node that needs to be replaced
	override, ok = overrides[block.Name]
	if !ok {
		goto end
	}
	// We found a block with an override - replace it
	// But we need to preserve the block's indentation!

	// Find indentation by checking previous nodes or the block itself
	indent = getIndentation(nodes, index)

	// Apply indentation to all lines of the override content
	override = applyIndentationToNodes(override, indent)

	// Replace the block node with its override
	nodes[index] = &OverrideNode{
		name:  block.Name,
		Elems: override,
	}
end:
	return err
}

// applyOverrides replaces block nodes in a template with override content.
// It traverses the template's node tree and replaces BlockNodes with their
// overridden content when found.
func (n *ParentNode) applyOverrides(t *Template, nodes []Node, overrides Overrides) (err error) {
	var kids []Node

	for i, node := range nodes {
		err = n.applyOverride(t, nodes, i, overrides)
		if err != nil {
			goto end
		}

		// Then recursively process any child nodes
		// This handles blocks inside sections, for example
		switch getter := node.(type) {
		case ChildrenGetter:
			kids = getter.GetChildren()
		case DerivedNodesGetter:
			kids, err = getter.GetDerivedNodes(t)
			if err != nil {
				goto end
			}
		}
		if kids == nil {
			continue
		}
		// TODO Try this without the .Clone() to see if it eliminates the need to update t.partials[].Elems = kids
		err = n.applyOverrides(t.Clone(), kids, overrides)
	}
end:
	return err
}

// getIndentation determines the indentation that should be applied to a block's content.
// This looks at surrounding nodes to figure out how much whitespace precedes the block.
// By default, no indentation
func getIndentation(nodes []Node, blockIndex int) (indent string) {
	var textNode TextNode
	var ok bool

	// Check if there's a TextNode right before this block with whitespace at the end
	if blockIndex <= 0 {
		goto end
	}
	textNode, ok = nodes[blockIndex-1].(TextNode)
	if !ok {
		goto end
	}
	// Extract trailing whitespace from the text node
	indent = extractTrailingIndent(textNode)

	// You might also check the block node itself if it stores indentation information
end:
	return indent
}

// extractTrailingIndent gets the whitespace at the end of a string
// after the last newline, which represents the indentation level.
func extractTrailingIndent(text TextNode) (indent string) {
	var trailing TextNode

	// Find the last newline
	lastNewline := -1
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] != '\n' {
			continue
		}
		lastNewline = i
		break
	}

	if lastNewline == -1 {
		// No newline, check if the entire text is whitespace
		if !isAllWhitespace(text) {
			goto end
		}
		indent = string(text)
		goto end
	}

	// Extract whitespace after the last newline
	trailing = text[lastNewline+1:]
	if !isAllWhitespace(trailing) {
		goto end
	}
	indent = string(trailing)
end:
	return indent
}

// isAllWhitespace checks if a string contains only whitespace characters.
func isAllWhitespace(text TextNode) bool {
	for _, c := range text {
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			return false
		}
	}
	return true
}

// applyIndentationToNodes adds indentation to every line of content in each node.
func applyIndentationToNodes(nodes []Node, indent string) (result []Node) {
	if indent == "" {
		result = nodes
		goto end
	}

	// For simplicity, convert nodes to text and back with indentation
	// In a real implementation, you'd want to handle this more efficiently
	result = make([]Node, len(nodes))

	for i, node := range nodes {
		textNode, ok := node.(TextNode)
		if !ok {
			// For other nodes, clone and keep them as is
			// Note: For complex content, you might need more sophisticated handling
			result[i] = node.Clone()
			continue
		}
		// For text nodes, add indentation to each line
		result[i] = TextNode(applyIndentationToText(textNode, indent))
	}
end:
	return result
}

// applyIndentationToText adds the given indentation to every line of text
// except the first line (which is already indented by its position).
func applyIndentationToText(text TextNode, indent string) string {
	// This is a simplified implementation
	// You'll need to handle newlines properly
	// and ensure only content lines get indented
	noop(indent)
	// Replace newlines with newline+indent
	// But avoid adding indentation after the final newline if present
	return string(text) // This is a placeholder - implement the actual indentation logic
}

// applyIndentationToChildNodes applies indentation to nodes derived from inheritance
func (pp *SpecConformance) applyIndentationToChildNodes(nodes []Node, indent string) {
	if indent == "" || len(nodes) == 0 {
		return
	}

	// Apply indentation to appropriate nodes
	for i, node := range nodes {
		textNode, ok := node.(TextNode)
		if !ok {
			continue
		}

		// Apply indentation to text nodes
		text := string(textNode)
		if !strings.Contains(text, "\n") {
			continue
		}

		// Add indentation after each newline
		lines := strings.Split(text, "\n")
		for j := 1; j < len(lines); j++ {
			if lines[j] != "" {
				lines[j] = indent + lines[j]
			}
		}
		nodes[i] = TextNode(strings.Join(lines, "\n"))
	}
}

// Name returns the value of the name property of ParentNode
func (n *ParentNode) Name() string {
	return n.name
}

// SetStandalone implements the StandaloneNode interface for ParentNode
func (n *ParentNode) SetStandalone(indent string) {
	n.isStandalone = true
	n.indent = indent
}

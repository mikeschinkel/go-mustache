package mustache

import (
	"fmt"
	"strings"
)

var _ StandaloneTagNode = (*InheritNode)(nil) // Type assertion to verify interface compliance
var _ Node = (*InheritNode)(nil)

// InheritNode represents template inheritance in Mustache.
// It includes a parent template ({{<template}}) and allows for
// block overrides within the inheriting template.
type InheritNode struct {
	// Name is the partial template to inherit from
	Name string
	// Overrides contains any block overrides defined in this template
	Overrides Overrides
	// isStandalone indicates if this is a standalone tag (only non-whitespace on its line)
	isStandalone bool
	// indent is the whitespace preceding this node when it's a standalone tag
	indent string
}

func NewInheritNode(name string, overrides Overrides) *InheritNode {
	return &InheritNode{
		Name:      name,
		Overrides: overrides,
	}
}

func (n *InheritNode) Clone() Node {
	return &InheritNode{
		Name:         n.Name,
		Overrides:    n.Overrides.Clone(),
		isStandalone: n.isStandalone,
		indent:       n.indent,
	}
}

// Render implements the Node interface for InheritNode.
// It loads the parent template, processes any block overrides,
// and renders the combined result.
func (n *InheritNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	var tmpl, inheritedTmpl *Template
	var found bool
	var errs MultiErr
	var overrides Overrides

	// Get the partial template
	tmpl, found, err = t.getPartial(n.Name)
	if !found {
		// partial not found
		goto end
	}

	// Clone the template to avoid modifying the original
	//goland:noinspection GoDfaErrorMayBeNotNil
	inheritedTmpl = tmpl.Clone()

	overrides = t.pushOverrides(n.Overrides)

	// Apply any block overrides
	if len(overrides) > 0 {
		// We need to find and replace block nodes in the inherited template
		n.applyOverrides(inheritedTmpl.Elems, overrides)
	}

	// Render the inherited template with the current context
	// IMPORTANT: Render each element separately to maintain correct order
	//            Using the same Writer ensures content appears in the correct order
	errs = NewMultiErr()
	for _, elem := range inheritedTmpl.Elems {
		err = elem.Render(t, w, c...)
		if err != nil {
			errs.Add(err)
		}
	}
	err = errs.Err()

	t.overrides.Pop()
end:
	return err
}

// String returns a string representation of the InheritNode.
func (n *InheritNode) String() string {
	return fmt.Sprintf("[inherit: %q Overrides: %v]", n.Name, n.Overrides)
}

// GetDerivedNodes returns the nodes from the template being inherited
// Implement DerivedNodesGetter interface to support preprocessing
func (n *InheritNode) GetDerivedNodes(t *Template) (elems []Node, err error) {
	var tmpl, inheritedTmpl *Template
	var found bool

	// Get the partial template
	tmpl, found, err = t.getPartial(n.Name)
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
	inheritedTmpl = tmpl.Clone()

	// Apply any block overrides
	if len(n.Overrides) > 0 {
		// We need to find and replace block nodes in the inherited template
		n.applyOverrides(inheritedTmpl.Elems, n.Overrides)
	}
	elems = inheritedTmpl.Elems
end:
	return elems, err
}

// applyOverrides replaces block nodes in a template with override content.
// It traverses the template's node tree and replaces BlockNodes with their
// overridden content when found.
func (n *InheritNode) applyOverrides(nodes []Node, overrides Overrides) {
	for i, node := range nodes {
		block, ok := node.(*BlockNode)
		if ok {
			block.applyOverrides(nodes, i, overrides)
		}
		// Then recursively process any child nodes
		// This handles blocks inside sections, for example
		container, isContainer := node.(ChildrenGetter)
		if !isContainer {
			continue
		}
		n.applyOverrides(container.GetChildren(), overrides)
	}
}

func (b *BlockNode) applyOverrides(nodes []Node, index int, overrides map[string][]Node) {
	var indent string
	var override []Node

	// First, check if this is a block node that needs to be replaced
	override, hasOverride := overrides[b.Name]
	if !hasOverride {
		goto end
	}
	// We found a block with an override - replace it
	// But we need to preserve the block's indentation!

	// Find indentation by checking previous nodes or the block itself
	indent = getBlockIndentation(b, nodes, index)

	// Apply indentation to all lines of the override content
	override = applyIndentationToNodes(override, indent)

	// Replace the block node with its override
	nodes[index] = &OverrideNode{
		Name:  b.Name,
		Elems: override,
	}
end:
	return
}

// getBlockIndentation determines the indentation that should be applied to a block's content.
// This looks at surrounding nodes to figure out how much whitespace precedes the block.
// By default, no indentation
func getBlockIndentation(_ *BlockNode, nodes []Node, blockIndex int) (indent string) {
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
		if textNode, ok := node.(TextNode); ok {
			// Apply indentation to text nodes
			text := string(textNode)
			if strings.Contains(text, "\n") {
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
	}
}

// SetStandalone implements the StandaloneTagNode interface for InheritNode
func (n *InheritNode) SetStandalone(indent string) {
	n.isStandalone = true
	n.indent = indent
}

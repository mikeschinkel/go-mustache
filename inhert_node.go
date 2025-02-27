package mustache

import (
	"fmt"
)

var _ Node = (*InheritNode)(nil)

// InheritNode represents template inheritance in Mustache.
// It includes a parent template ({{<template}}) and allows for
// block overrides within the inheriting template.
type InheritNode struct {
	// Name is the partial template to inherit from
	Name string
	// Overrides contains any block overrides defined in this template
	Overrides map[string][]Node
}

func (n *InheritNode) Clone() Node {
	newNode := new(InheritNode)
	*newNode = *n
	return newNode
}

// Render implements the Node interface for InheritNode.
// It loads the parent template, processes any block overrides,
// and renders the combined result.
func (n *InheritNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	var ok bool
	var tmpl, inheritedTmpl *Template

	w.tag()
	//defer w.tag()

	// Get the partial template
	tmpl, ok = t.partials[n.Name]
	if !ok {
		if !t.silentMiss {
			err = fmt.Errorf("template '%s' not found", n.Name)
		}
		goto end
	}

	// Clone the template to avoid modifying the original
	inheritedTmpl = tmpl.Clone()

	// Apply any block overrides
	if len(n.Overrides) > 0 {
		// We need to find and replace block nodes in the inherited template
		applyOverrides(inheritedTmpl, n.Overrides)
	}

	// Render the inherited template with the current context
	err = inheritedTmpl.Render(w.w, c...)
end:
	return err
}

// String returns a string representation of the InheritNode.
func (n *InheritNode) String() string {
	return fmt.Sprintf("[inherit: %q Overrides: %v]", n.Name, n.Overrides)
}

// applyOverrides replaces block nodes in a template with override content.
// It traverses the template's node tree and replaces BlockNodes with their
// overridden content when found.
func applyOverrides(tmpl *Template, overrides map[string][]Node) {
	// Start by recursively processing the top-level nodes
	replaceBlocksInNodes(tmpl.Elems, overrides)
}

// replaceBlocksInNodes recursively walks through a slice of nodes,
// replacing any BlockNodes with their overridden content if available.
func replaceBlocksInNodes(nodes []Node, overrides map[string][]Node) {
	for i, node := range nodes {
		// First, check if this is a block node that needs to be replaced
		if block, ok := node.(*BlockNode); ok {
			if override, hasOverride := overrides[block.Name]; hasOverride {
				// We found a block with an override - replace it
				// But we need to preserve the block's indentation!

				// Find indentation by checking previous nodes or the block itself
				indent := getBlockIndentation(block, nodes, i)

				// Apply indentation to all lines of the override content
				indentedOverride := applyIndentationToNodes(override, indent)

				// Replace the block node with its override
				nodes[i] = &OverrideNode{
					Name:  block.Name,
					Elems: indentedOverride,
				}
			}
		}

		// Then recursively process any child nodes
		// This handles blocks inside sections, for example
		if container, isContainer := node.(ChildrenGetter); isContainer {
			replaceBlocksInNodes(container.GetChildren(), overrides)
		}
	}
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
	noop()
	// Replace newlines with newline+indent
	// But avoid adding indentation after the final newline if present
	return string(text) // This is a placeholder - implement the actual indentation logic
}

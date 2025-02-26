package mustache

// Node is the interface that represents a node in the template parse tree.
// Every parsed element in a template (variables, sections, text, etc.) is
// represented as a Node implementation.
//
// The Node interface is the core abstraction in the template system. The parser
// produces a tree of nodes, which can then be rendered with different contexts.
// Each specific node type (VarNode, SectionNode, TextNode, etc.) implements this
// interface with its own rendering logic.
type Node interface {
	// Render evaluates the node within the given template and context,
	// writing its output to the provided Writer.
	//
	// Parameters:
	//   - t: The template containing this node
	//   - w: The writer to output the rendered result
	//   - c: One or more context objects providing the data for rendering
	//
	// Returns an error if rendering fails.
	Render(t *Template, w *Writer, c ...interface{}) error
}

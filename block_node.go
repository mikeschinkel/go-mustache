package mustache

import (
	"fmt"
)

var _ Node = (*BlockNode)(nil)
var _ ChildrenGetter = (*BlockNode)(nil)

// BlockNode represents a block definition in Mustache inheritance.
// It defines a block of content that can be overridden in child templates.
// The syntax is {{$blockName}}default content{{/blockName}}.
type BlockNode struct {
	// Name is the block identifier
	Name string
	// Elems contains the default content nodes inside the block
	Elems []Node
}

func (n *BlockNode) Clone() Node {
	newNode := new(BlockNode)
	*newNode = *n
	newNode.Elems = cloneNodes(n.Elems)
	return newNode
}

// GetChildren returns the block's child nodes.
// This method implements the ChildrenGetter interface.
func (n *BlockNode) GetChildren() []Node {
	return n.Elems
}

// Render implements the Node interface for BlockNode.
// It renders the default content of the block.
func (n *BlockNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	w.tag()
	//defer w.tag()
	errs := MultiErr{}
	//n.Elems = fixWhitespace(n.Elems, nil)
	for _, elem := range n.Elems {
		err = elem.Render(t, w, c...)
		errs.Add(t.triageError(w.w,
			fmt.Errorf("failed to render block element %s; %w", n.Name, err),
		))
	}
	return errs.Err()
}

// String returns a string representation of the BlockNode for debugging.
func (n *BlockNode) String() string {
	return fmt.Sprintf("[block: %q Elems: %s]", n.Name, n.Elems)
}

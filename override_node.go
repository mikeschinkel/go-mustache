package mustache

import (
	"fmt"
)

// Ensure OverrideNode implements StandaloneNode
var _ StandaloneNode = (*OverrideNode)(nil)

func (o *OverrideNode) SetStandalone(indent string) {
	// Set this node as standalone with given indent
	o.isStandalone = true
	o.indent = indent
}

func (o *OverrideNode) Name() string {
	return o.name // Return the block name
}

func (o *OverrideNode) GetDerivedNodes(_ *Template) ([]Node, error) {
	return o.Elems, nil
}

// OverrideNode represents a block that has been overridden.
// It behaves like a BlockNode but contains the override content.
type OverrideNode struct {
	name         string
	isStandalone bool
	indent       string
	Elems        []Node
}

func (n *OverrideNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.tag()
	//defer w.tag()

	for _, elem := range n.Elems {
		if err := elem.Render(t, w, c...); err != nil {
			return err
		}
	}

	return nil
}

func (n *OverrideNode) Clone() Node {
	clone := &OverrideNode{
		name:  n.name,
		Elems: make([]Node, len(n.Elems)),
	}

	for i, elem := range n.Elems {
		clone.Elems[i] = elem.Clone()
	}

	return clone
}

func (n *OverrideNode) String() string {
	return fmt.Sprintf("[override: %q Elems: %v]", n.name, n.Elems)
}

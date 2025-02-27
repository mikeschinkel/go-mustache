package mustache

import (
	"fmt"
)

// OverrideNode represents a block that has been overridden.
// It behaves like a BlockNode but contains the override content.
type OverrideNode struct {
	Name  string
	Elems []Node
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
		Name:  n.Name,
		Elems: make([]Node, len(n.Elems)),
	}

	for i, elem := range n.Elems {
		clone.Elems[i] = elem.Clone()
	}

	return clone
}

func (n *OverrideNode) String() string {
	return fmt.Sprintf("[override: %q Elems: %v]", n.Name, n.Elems)
}

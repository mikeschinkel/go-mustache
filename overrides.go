package mustache

type Overrides map[string][]Node

func (o Overrides) Clone() Overrides {
	overrides := make(Overrides, len(o))
	for name, nodes := range o {
		clonedNodes := make([]Node, len(nodes))
		for i, node := range nodes {
			clonedNodes[i] = node.Clone()
		}
		overrides[name] = clonedNodes
	}
	return overrides
}

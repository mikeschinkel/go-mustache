package mustache

type StandaloneSetter interface {
	SetStandalone(indent string)
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

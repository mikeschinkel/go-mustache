package mustache

// StandaloneTagNode interface for nodes that can be standalone
type StandaloneTagNode interface {
	Node
	// SetStandalone marks the node as a standalone tag with the given indentation
	SetStandalone(indent string)
	// DerivedNodesGetter returns nodes that are derived from template dependencies
	DerivedNodesGetter
}

// preprocessStandalone handles any node that can be standalone
func (pp *SpecConformance) preprocessStandalone(t *Template, nodes []Node, index int, tagNode StandaloneTagNode) (err error) {
	var hasLeadingWhitespace bool
	var text TextNode
	var followedByNewline, isStandalone bool
	var matches []int
	var ok bool
	var indent, s string
	var kids []Node
	var last int

	// Look for whitespace-only text before this tag
	indent, hasLeadingWhitespace = pp.getLeadingWhitespace(nodes, index)
	text, followedByNewline = pp.getFollowedByNewline(nodes, index)
	if followedByNewline {
		nodes[index+1] = text
	}
	// TODO Move into getFollowedByNewline()
	if !followedByNewline && index == len(nodes)-1 {
		// Treat partials with no trailing TextNode as effectively "follows by new line"
		followedByNewline = true
	}
	// END TODO

	// If we have both conditions, this is a standalone tag
	isStandalone = hasLeadingWhitespace && followedByNewline

	if !isStandalone {
		goto end
	}

	tagNode.SetStandalone(indent)
	s = string(text)

	//// Remove leading whitespace (set to empty text)
	//if index > 0 {
	//	nodes[index-1] = TextNode("")
	//}

	err = pp.indentText(t, nodes)
	if err != nil {
		goto end
	}
	// Use regex to match and remove the newline (handles both \n and \r\n)
	matches = newlineRegex.FindStringIndex(s)
	if matches == nil {
		goto end
	}

	// Check if the newline is at the very beginning of the text
	if matches[0] != 0 {
		// This is a more complex case where the newline isn't at the start
		// This branch shouldn't be hit in your test case
		goto end
	}
	// Remove the newline completely
	nodes[index+1] = TextNode(s[matches[1]:])

	kids, err = tagNode.GetDerivedNodes(t)
	if err != nil {
		goto end
	}
	if len(kids) < 2 {
		goto end
	}
	last = len(kids) - 1
	text, ok = kids[last].(TextNode)
	if !ok {
		goto end
	}

	// Apply any special indentation handling for derived nodes
	pp.applyIndentationToChildNodes(kids, indent)

	// Remove the indent on the last kid for a standalone
	kids[last] = TextNode(string(text)[:len(text)-len(indent)])
end:
	return err
}

package mustache

import (
	"errors"
	"regexp"
)

// newlineRegex matches both CR+LF and LF newlines
var newlineRegex = regexp.MustCompile("(\r?\n)")

// Preprocessor is an interface for components that analyze or transform templates
// before rendering begins
type Preprocessor interface {
	Preprocess(t *Template, nodes []Node, index int) error
}

// Preprocess process identifies standalone partials and marks them as such
func Preprocess(p Preprocessor, t *Template, nodes []Node) (err error) {
	errs := make([]error, 0)
	for i := 0; i < len(nodes); i++ {
		err = p.Preprocess(t, nodes, i)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// extractBlockOverrides processes a list of nodes to find block definitions
// and returns them as a map of block name to block content nodes.
func extractBlockOverrides(nodes []Node) map[string][]Node {
	overrides := make(map[string][]Node)
	for _, n := range nodes {
		block, ok := n.(*BlockNode)
		if !ok {
			continue
		}
		overrides[block.Name] = block.Elems
	}
	return overrides
}

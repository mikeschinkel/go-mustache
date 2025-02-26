package mustache

import (
	"errors"
)

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

package mustache

import (
	"errors"
	"fmt"
	"strings"
)

var _ DerivedNodesGetter = (*PartialNode)(nil)

// The PartialNode type represents a named partial template.
type PartialNode struct {
	name         string
	isDynamic    bool
	isStandalone bool
	indent       string
}

func (p *PartialNode) GetDerivedNodes(t *Template) (nodes []Node, err error) {
	tmpl, err := p.getPartialsTemplate(t)
	if err != nil {
		goto end
	}
	nodes = tmpl.Elems
end:
	return nodes, err
}

func (p *PartialNode) getPartialsTemplate(t *Template) (tmpl *Template, err error) {
	var ok bool
	var root map[string]any
	var value any

	name := p.name

	// If this is a dynamic partial, resolve its actual name
	if !p.isDynamic {
		print()
	} else {
		root, ok = t.root.(map[string]any)
		if !ok {
			err = fmt.Errorf("unexpected coding error for '%s'; root not a map[string]any", p.name)
			goto end
		}
		value, err = p.getNestedValue(root, name, t.partials)
		if err != nil {
			// per the spec, a failed lookup should result in an empty string
			// rather than an error for dynamic partials
			if !t.silentMiss {
				// But if the user is requesting to see the error, throw it
				err = errors.Join(ErrDynamicRefNotFound,
					fmt.Errorf("partial_name=%s", p.name),
					err,
				)
			}
			goto end
		}
		switch vt := value.(type) {
		case string:
			name = vt
		default:
			goto end
		}
	}

	// Look up the partial template
	tmpl, ok = t.partials[name]
	if !ok {
		err = fmt.Errorf("partial node '%s' not found", p.name)
		goto end
	}
end:
	return tmpl, err
}

func (p *PartialNode) Render(t *Template, w *Writer, c ...interface{}) (err error) {
	var tmpl *Template

	w.tag()

	tmpl, err = p.getPartialsTemplate(t)
	if err != nil {
		err = fmt.Errorf("partial node '%s' not found", p.name)
		goto end
	}

	tmpl.partials = t.partials
	tmpl.root = t.root
	err = tmpl.render(w, c...)

end:
	return err
}
func (p *PartialNode) String() string {
	return fmt.Sprintf("[partial: %s]", p.name)
}

//goland:noinspection GoUnusedParameter
func (p *PartialNode) getNestedValue(data map[string]any, path string, partials map[string]*Template) (value any, err error) {
	var key string
	var i int
	var nextMap map[string]any
	var exists, ok bool
	var isDynamic bool
	var s string

	// Split the path into individual keys
	keys := strings.Split(path, ".")

	// Start with the initial map
	current := data

	// Traverse through each key in the path
	for i, key = range keys {
		isDynamic = strings.HasPrefix(key, "*")
		if isDynamic {
			key = key[1:] // Remove the * prefix
		}

		// Check if we're at the final key
		// Try to get the next nested map
		value, exists = current[key]
		if !exists {
			err = fmt.Errorf("key '%s' not found for '%s'", key, path)
			goto end
		}

		if isDynamic {
			s, ok = value.(string)
			if !ok {
				err = fmt.Errorf("dynamic reference '%s' must resolve to string", key)
				goto end
			}
			value = s
		}

		if i == len(keys)-1 {
			goto end
		}

		nextKey := keys[i+1]
		if strings.HasPrefix(nextKey, "*") {
			s, ok = value.(string)
			if !ok {
				err = fmt.Errorf("value at '%s' must resolve to string for dynamic lookup", key)
				goto end
			}

			value, exists = current[s]
			if !exists {
				err = fmt.Errorf("resolved key '%s' not found", s)
				goto end
			}

			nextMap, ok = value.(map[string]any)
			if !ok {
				err = fmt.Errorf("resolved value at '%s' is not a map", s)
				goto end
			}
			current = nextMap
			continue
		}

		nextMap, ok = value.(map[string]any)
		if !ok {
			err = fmt.Errorf("value at '%s' within '%s' is not a map", key, path)
			value = nil
			goto end
		}
		current = nextMap
	}
end:
	return value, err
}

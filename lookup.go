// Copyright (c) 2025 Mike Schinkel
// Portions Copyright (c) 2014 Alex Keratitis
// Portions Copyright (c) 2009 Michael Hoisie

package mustache

import (
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Lookup searches for a value with the given name in the provided context chain.
// It supports dot notation for nested lookups (e.g., "user.name") and implements
// the Mustache context lookup rules.
//
// The lookup follows this precedence order:
// 1. Methods on the context object
// 2. Struct fields with matching tags
// 3. Struct fields with matching names
// 4. Map keys with matching names
//
// Parameters:
//   - name: The name/path to look up
//   - context: One or more context objects to search in (first match wins)
//
// Returns:
//   - The found value (or nil if not found)
//   - A boolean indicating if the value was found
//   - An error if the lookup failed
//
// For dot notation, the lookup is performed recursively, first resolving
// the left side of the dot, then using that result as context for the right side.
func (t *Template) Lookup(name string, context ...interface{}) (interface{}, bool, error) {
	// If the dot notation was used we split the word in two and perform two
	// consecutive lookups. If the first one fails we return no value and a
	// negative Truthiness. Taken from github.com/hoisie/mustache.
	if name != "." && strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		value, ok, err := t.Lookup(parts[0], context...)
		if !ok || err != nil {
			return nil, false, err
		}
		return t.Lookup(parts[1], value)
	}
	// Iterate over the context chain and try to match the name to a value.
	for _, c := range context {
		// Reflect on the value and type of the current context.
		reflectValue := reflect.ValueOf(c)
		if reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		// If the name is ".", we should return the whole context as-is.
		if name == "." {
			return c, Truthiness(reflectValue), nil
		}
		kind := reflectValue.Kind()
		//goland:noinspection GoSwitchMissingCasesForIotaConsts
		switch kind {
		case reflect.Map:
			// If the current context is a map, we'll look for a key in that map
			// that matches the name.
			//
			// Try to match a map key to the name. For example:
			//
			// 	m := map[string]string{"foo": "bar"}
			//  mustache.Render("{{foo}}", m)
			item := reflectValue.MapIndex(reflect.ValueOf(name))
			if item.IsValid() {
				return item.Interface(), Truthiness(item), nil
			}
		case reflect.Struct:
			// If the current context is a struct
			// First try to match against a method. This allows overriding everything.
			field, found := getFieldByMethod(reflectValue, name)
			if !found {
				// Try title-casing the name in case the name in the template was written in
				// lowercase.
				titler := cases.Title(language.English)
				field, found = getFieldByMethod(reflectValue, titler.String(name))
			}
			fldKind := field.Kind()
			if fldKind == reflect.Invalid {
				// If no method was matched, we'll try to match a tag. This is
				// useful for matching fields that have a different name than the
				// one we want to use in our templates. For example:
				//
				// 	type Foo struct {
				// 		Bar string `template:"baz"`
				// 	}
				//  ctx := &Foo{"qux"}
				//  mustache.Render("{{baz}}", ctx)
				//
				field, found = t.getFieldByTag(reflectValue, name)
				fldKind = field.Kind()
			}
			if fldKind == reflect.Invalid {
				// If no other matches, we'll try to match by Go struct property name
				// For example:
				//
				// 	type Foo struct { Bar string }
				//  ctx := &Foo{"baz"}
				//  mustache.Render("{{Bar}}", ctx)
				field = reflectValue.FieldByName(name)
				fldKind = field.Kind()
			}
			if fldKind == reflect.Invalid {
				return nil, false, fmt.Errorf("failed to Lookup field by tag %s", name)
			}
			if fldKind == reflect.Ptr {
				field = field.Elem()
			}
			if field.IsValid() {
				return field.Interface(), Truthiness(field), nil
			}
			return nil, false, nil
		}
		// If by this point no value was matched, we'll move up a step in the
		// chain and try to match a value there.
	}
	// We've exhausted the whole context chain and found nothing. Return a nil
	// value and a negative Truthiness.
	return nil, false, nil
}

// The Truthiness function will tell us if r is a truthy value or not. This is
// important for sections as they will render their content based on the output
// of this function.
//
// Zero values are considered falsy. For example an empty string, the integer 0
// and so on are all considered falsy.
func Truthiness(r reflect.Value) bool {
out:
	switch r.Kind() {
	case reflect.Array, reflect.Slice:
		return r.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return r.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return r.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return r.Float() > 0
	case reflect.String:
		return r.String() != ""
	case reflect.Bool:
		return r.Bool()
	case reflect.Ptr, reflect.Interface:
		r = r.Elem()
		goto out
	default:
		if !r.IsValid() {
			return false
		}
		return r.Interface() != nil
	}
}

// getFieldByTag uses the tag to Lookup the field vs. relying on the property name
// of the Go struct used in the view object.
func (t *Template) getFieldByTag(rv reflect.Value, tagValue string) (fld reflect.Value, found bool) {
	rt := rv.Type()
	key := rt.String() + "|" + tagValue

	// If it's a pointer, get the underlying value and type
	if rt.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	tagIndex, ok := t.tagIndexCache[key]
	if !ok && tagIndex == 0 {
		tagIndex = -1
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			tag := ft.Tag.Get(t.structTag)
			if tag != tagValue {
				continue
			}
			tagIndex = i
			break
		}
	}
	if tagIndex == -1 {
		fld = reflect.Value{}
		t.tagIndexCache[key] = -1
		goto end
	}
	fld = rv.Field(tagIndex)
	found = fld.IsValid()
	t.tagIndexCache[key] = tagIndex
end:
	// If it's a pointer, get the underlying value
	if fld.Kind() == reflect.Ptr {
		fld = fld.Elem()
	}
	return fld, found
}

// getFieldByMethod calls a method to determine the field name vs. relying on the field name in the mustache file to match the field name
func getFieldByMethod(rv reflect.Value, methodName string) (fld reflect.Value, found bool) {
	var rt reflect.Type
	var method reflect.Value

	// If we start with a non-pointer, try getting the method directly first
	if rv.Kind() != reflect.Ptr {
		method = rv.MethodByName(methodName)

		// If not found, try getting it through a pointer by creating a new value
		if !method.IsValid() {
			newVal := reflect.New(rv.Type())
			newVal.Elem().Set(rv)
			method = newVal.MethodByName(methodName)
		}
	} else {
		// If we start with a pointer, try pointer method first
		method = rv.MethodByName(methodName)

		// If not found, try the non-pointer method
		if !method.IsValid() && !rv.IsNil() {
			method = rv.Elem().MethodByName(methodName)
		}
	}

	if !method.IsValid() {
		goto end
	}

	rt = method.Type()
	if rt.NumIn() != 0 {
		goto end
	}

	fld = method.Call(nil)[0]
	found = Truthiness(fld)

end:
	// If it's a pointer, get the underlying value
	if fld.Kind() == reflect.Ptr {
		fld = fld.Elem()
	}
	return fld, found
}

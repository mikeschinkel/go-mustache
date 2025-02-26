package mustache

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
)

// injectError writes an error message directly into the output.
// This is used when SilentMiss and InjectOnMiss are both true,
// to provide visual feedback about missing variables in the output.
func injectError(w io.Writer, err error) {
	_, err = fmt.Fprintf(w, "[ERROR: %s]", err.Error())
	if err != nil {
		panic(err)
	}
}

// escape performs HTML escaping of special characters in accordance with the Mustache spec.
// This function is similar to text/template.HTMLEscapeString but preserves
// "&apos;" and "&quot;" for compatibility with the Mustache specification.
//
// The following characters are escaped:
// - " becomes &quot;
// - ' becomes &apos;
// - & becomes &amp;
// - < becomes &lt;
// - > becomes &gt;
func escape(s string) string {
	if strings.IndexAny(s, `'"&<>`) < 0 {
		return s
	}
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// write formats and writes a value to the provided writer.
// It handles different types of values appropriately:
// - Stringers: uses their String() method
// - strings: writes with %s format
// - integers: writes with %d format
// - floats: writes with %g format (compact representation)
// - other types: writes with %v format (default representation)
//
// This ensures that values are rendered with the most appropriate formatting
// based on their type.
func write(w io.Writer, v interface{}) {
	if s, ok := v.(fmt.Stringer); ok {
		_, _ = fmt.Fprint(w, s.String())
	} else {
		switch v.(type) {
		case string:
			MustFprintf(w, "%s", v)
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			MustFprintf(w, "%d", v)
		case float32, float64:
			MustFprintf(w, "%g", v)
		default:
			MustFprintf(w, "%v", v)
		}
	}
}

// MustFprintf is a wrapper around fmt.Fprintf that logs errors instead of returning them.
// This is used throughout the package to simplify error handling when writing to output.
// If writing fails, it logs the error but allows processing to continue.
func MustFprintf(w io.Writer, format string, a ...interface{}) (n int) {
	var err error
	n, err = fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Print("ERROR: FPrintf failed to write to file")
	}
	return n
}

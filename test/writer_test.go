package mustache_test

import (
	"bytes"
	"testing"

	"github.com/alexkappa/mustache"
)

func TestWriter(t *testing.T) {
	for _, test := range []struct {
		text     bool
		tag      bool
		input    string
		expected string
	}{
		{true, true, "some text\n", "some text\n"},
		{false, true, "  {{#standalone}}\n here.", " here."},
		{false, false, "print this\n and this", "print this\n and this"},
	} {
		b := bytes.NewBuffer(nil)
		w := mustache.NewWriter(b)
		w.HasText = test.text
		w.HasTag = test.tag
		for _, r := range test.input {
			err := w.WriteRune(r)
			if err != nil {
				t.Errorf("write error %q", err)
			}
		}
		must("flush", w.Flush())
		if b.String() != test.expected {
			t.Errorf("unexpected output %q, expected %q", b.String(), test.expected)
		}
	}
}

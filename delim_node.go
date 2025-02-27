package mustache

var _ Node = (*DelimNode)(nil)

type DelimNode string

func (n DelimNode) Clone() Node {
	return n
}

func (n DelimNode) String() string {
	return "[delim]"
}

//goland:noinspection GoUnusedParameter
func (n DelimNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.tag()
	return nil
}

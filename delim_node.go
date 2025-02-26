package mustache

type DelimNode string

func (n DelimNode) String() string {
	return "[delim]"
}

//goland:noinspection GoUnusedParameter
func (n DelimNode) Render(t *Template, w *Writer, c ...interface{}) error {
	w.tag()
	return nil
}

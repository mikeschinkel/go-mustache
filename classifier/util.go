package classifier

//goland:noinspection GoUnusedParameter
func noop(args ...any) {}

func isInlineWhitespace(c byte) (is bool) {
	return c == ' ' || c == '\t' || c == '\r'
}

package classifier

//goland:noinspection GoUnusedParameter,GoUnusedFunction
func noop(args ...any) {}

func isInlineWhitespace(c byte) (is bool) {
	return c == ' ' || c == '\t' || c == '\r'
}

// addToPtr returns a pointer to a value of type T that contains the value at *T incremented by incr
func addToPtr[T numeric](pos *T, incr T) *T {
	newPos := *pos
	newPos += incr
	return &newPos
}

package classify

// isInlineWhitespace determines if a character is considered "inline whitespace"
// which this project defines as spaces, tabs, and carriage returns, but not
// newlines.
func isInlineWhitespace(c byte) (is bool) {
	return c == ' ' || c == '\t' || c == '\r'
}

// todo is a placeholder function that does nothing. It is useful for when code
// that is not yet working needs commented out in order to test other parts of
// the code, but the commented-out code and especially declared variables that
// code depends on needs to bot be commented out. e.g.:
//
//	var foo int
//	// doSomething(foo)  // temporarily comment this out
//	todo(foo)	           // temporarily add this like adding a TODO comment
//
// This is better than just commenting as there is no good reason so use todo()
// in production code making it easy to ensure all code commented out for testing
// and debugging is properly addressed prior to releasing a new production
// version.
//
//goland:noinspection GoUnusedParameter,GoUnusedFunction
func todo(args ...any) {}

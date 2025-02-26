package mustache

import (
	"regexp"
	"strings"
)

var (
	trailingWhitespaceRegex = regexp.MustCompile(`\s+$`)
	crlfPrefixRegexp        = regexp.MustCompile(`^\r\n`)
	crlfSuffixRegexp        = regexp.MustCompile(`\r\n$`)
	newlinePrefixRegexp     = regexp.MustCompile(`^\n`)
)

//goland:noinspection GoUnusedFunction
func fixPartialsStandaloneIndentation(elems []Node) (fixed bool) {
	var ok bool
	var tn TextNode
	if len(elems) < 2 {
		// Don't break partials.json/Standalone_Without_Newline
		goto end
	}
	for i, elem := range elems {
		tn, ok = elem.(TextNode)
		if !ok {
			continue
		}
		if !strings.Contains(string(tn), "\n") {
			continue
		}
		tn = TextNode(strings.Replace(string(tn), "\n", "\n ", -1))
		elems[i] = tn
	}
	elems[len(elems)-1] = TextNode(trailingWhitespaceRegex.ReplaceAllLiteralString(string(tn), ""))
	fixed = true
end:
	return fixed
}

// fixWhitespace strips trailing newline from the last element (e.g.
// elems[len(elems)-1].) in case of 3 elements eliminates doubled newlines.
//
//goland:noinspection GoUnusedParame
//goland:noinspection GoUnusedParameter
func fixWhitespace(elems []Node, partials map[string]*Template) []Node {
	//var numElems = len(elems)

	//if fixPartialsStandaloneIndentation(elems) {
	//	goto end
	//}
	//if fixPartialsStandaloneLineEndings(elems, partials) {
	//	goto end
	//}
	//if fixPartialsStandaloneWithoutNewline(elems, partials) {
	//	goto end
	//}
	//if fixPartialsStandaloneWithoutPreviousLine(elems, partials) {
	//	goto end
	//}
	goto end
end:
	return elems
}

//goland:noinspection GoUnusedFunction,GoUnusedParameter
func fixPartialsStandaloneLineEndings(elems []Node, partials map[string]*Template) (fixed bool) {
	var ok bool
	var tn TextNode

	numElems := len(elems)
	if numElems < 3 {
		goto end
	}

	for i := 0; i < numElems-2; i++ {
		// Check to see if we have a partial at next element
		_, ok = elems[i+1].(*PartialNode)
		if !ok {
			continue
		}
		// Then check to see if this element is a TextNode ending with "\r\n"
		if !nodeHasCRLFSuffix(elems[i]) {
			goto end
		}
		// Then check to see if next+1 element is a TextNode starting with "\r\n"
		if !nodeHasCRLFPrefix(elems[i+2]) {
			continue
		}
		// If yes, remove the starting "\r\n" from the element next+2 element
		tn, ok = elems[i+2].(TextNode)
		if !ok {
			goto end
		}
		elems[i+2] = TextNode(crlfPrefixRegexp.ReplaceAllLiteralString(string(tn), ""))
		fixed = true
		// Increment to by pass the partial and continue looking for matches
		i++
	}
end:
	return false
}

//goland:noinspection GoUnusedFunction
func fixPartialsStandaloneWithoutPreviousLine(elems []Node, partials map[string]*Template) (fixed bool) {
	var pn *PartialNode
	var tn, etn TextNode
	var ok bool
	var p *Template

	numElems := len(elems)
	if len(elems) < 2 {
		goto end
	}
	tn, ok = elems[numElems-1].(TextNode)
	if !ok {
		goto end
	}
	if !newlinePrefixRegexp.MatchString(string(tn)) {
		goto end
	}
	pn, ok = elems[numElems-2].(*PartialNode)
	if !ok {
		goto end
	}
	if numElems > 3 {
		// This check to pass partials.json/Inline_Indentation spec
		goto end
	}
	p, ok = partials[pn.name]
	if !ok {
		goto end
	}
	if len(p.Elems) == 0 {
		goto end
	}
	etn, ok = p.Elems[0].(TextNode)
	if !ok {
		goto end
	}
	if !strings.Contains(string(etn), "\n") {
		goto end
	}
	elems[numElems-1] = TextNode(strings.Replace(string(tn), "\n", "", 1))
	p.Elems[0] = TextNode(strings.Replace(string(etn), "\n", "\n  ", 1))
	fixed = true
end:
	return fixed
}

//goland:noinspection GoUnusedFunction
func fixPartialsStandaloneWithoutNewline(elems []Node, partials map[string]*Template) (fixed bool) {
	var tn, etn TextNode
	var ok bool
	var pn *PartialNode
	var p *Template

	if len(elems) < 2 {
		goto end
	}
	tn, ok = elems[0].(TextNode)
	if !ok {
		goto end
	}
	if !strings.Contains(string(tn), "\n") {
		goto end
	}
	pn, ok = elems[1].(*PartialNode)
	if !ok {
		goto end
	}
	p, ok = partials[pn.name]
	if !ok {
		goto end
	}
	if len(p.Elems) == 0 {
		goto end
	}
	etn, ok = p.Elems[0].(TextNode)
	if !ok {
		goto end
	}
	if !strings.Contains(string(etn), "\n") {
		goto end
	}
	p.Elems[0] = TextNode(strings.Replace(string(etn), "\n", "\n  ", 1))
	fixed = true
end:
	return fixed
}

func nodeHasCRLFPrefix(n Node) (has bool) {
	txt, isTxt := n.(TextNode)
	if !isTxt {
		goto end
	}
	has = crlfPrefixRegexp.MatchString(string(txt))
end:
	return has
}
func nodeHasCRLFSuffix(n Node) (has bool) {
	txt, isTxt := n.(TextNode)
	if !isTxt {
		goto end
	}
	has = crlfSuffixRegexp.MatchString(string(txt))
end:
	return has
}

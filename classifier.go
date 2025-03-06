package mustache

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
)

//
//type detectorState struct {
//	lineStart   int
//	tagOpenPos  int
//	tagClosePos int
//}
//
//func newDetectorState() *detectorState {
//	return &detectorState{
//		tagOpenPos:  -1,
//		tagClosePos: -1,
//	}
//}

type TemplateClassifier struct {
	template   string
	openDelim  string
	closeDelim string
	lines      []string
	tagEndPos  int
	LineTypes  LineTypes
}

var crlfRegex = regexp.MustCompile("(\r\n|\r|\n)")

func NewTemplateClassifier(template string) *TemplateClassifier {
	lines := crlfRegex.Split(template, -1)
	return &TemplateClassifier{
		template:   template,
		openDelim:  "{{",
		closeDelim: "}}",
		lines:      lines,
	}
}

func (tc *TemplateClassifier) Initialize() (err error) {
	tc.LineTypes = make(LineTypes, len(tc.lines))
	tc.tagEndPos = -1
	return err
}

func (tc *TemplateClassifier) AddLineType(lineNo int, lts ...LineAspect) {
	ts := tc.LineTypes
	if lineNo >= len(ts) {
		panic(fmt.Sprintf("ERROR: LineTypes lineNo %d out of range [0..%d]", lineNo, len(ts)))
	}

	m := make(map[LineAspect]struct{}, len(ts[lineNo]))
	for _, t := range append(ts[lineNo], lts...) {
		m[t] = struct{}{}
	}
	ts[lineNo] = slices.Collect(maps.Keys(m))
	tc.LineTypes = ts
}

// Classify analyzes a mustache template string and returns
// a slice of booleans indicating whether each line contains a standalone tag.
// This implementation properly handles line boundaries and tag detection.
func (tc *TemplateClassifier) Classify() (_ LineTypes, err error) {
	errs := NewMultiErr()
	err = tc.Initialize()
	if err != nil {
		errs.Add(err)
		goto end
	}
	// Early return for empty template
	if len(tc.template) == 0 {
		tc.AddLineType(0, EmptyLine)
		goto end
	}

	// Process each line
	for i := 0; i < len(tc.lines); i++ {
		line := tc.lines[i]
		if line == "" {
			tc.AddLineType(i, EmptyLine)
			continue
		}
		// Check if this line contains exactly one standalone tag
		err = tc.classifyLine(&i)
		if err != nil {
			errs.Add(err)
		}
		noop()
	}
end:
	return tc.LineTypes, errs.Err()
}

const TripleBracketBegin = "{{{"
const TripleBracketEnd = "}}}"

func (tc *TemplateClassifier) classifyLine(lineNo *int) (err error) {
	var start int
	var tagOpened, tagClosed, multiline, notStandalone bool
	var lt LineAspect

	line := tc.lines[*lineNo]

	tc.tagEndPos = -1

	// Skip leading whitespace to find start of content
	pos := 0
	for pos < len(line) && tc.isNonNewlineWhitespace(line[pos]) {
		pos++
	}

	// If we've reached the end of the line, it's just whitespace, not a tag
	if pos >= len(line) {
		tc.AddLineType(*lineNo, WhitespaceLine)
		goto end
	}

	// Consume up to and including opening tag delimiter, if exists
	//
	// TODO We could optimize the following a bit by having two paths; one where we
	// have custom delimiter and one where do do not.
	for pos < len(line) {
		var ate bool
		ate, lt, err = tc.maybeEatTripleBracket(line, &pos)
		if err != nil {
			goto end
		}
		if ate {
			tc.AddLineType(*lineNo, TripleBraceUnescaped)
			tc.AddLineType(*lineNo, lt)
			goto end
		}
		if line[pos] != tc.openDelim[0] {
			// Not (yet?) at an opening delimiter
			notStandalone = true
			pos++
			continue
		}
		if line[pos+1] != tc.openDelim[1] {
			// 2nd char not matching a delimiter
			tc.AddLineType(*lineNo, TextLine)
			goto end
		}
		tagOpened = true
		pos += len(tc.openDelim) // Skip past the opening delimiter
		multiline = pos == len(line)
		break
	}

	if !tagOpened && pos >= len(line) {
		tc.AddLineType(*lineNo, TextLine)
		goto end
	}

	start = *lineNo

	tagClosed, line, err = tc.processTagPrefix(lineNo, &pos)
	if tagClosed && pos >= len(line) {
		if notStandalone {
			tc.AddLineType(*lineNo, InlineTagLine)
		} else {
			tc.AddLineType(*lineNo, StandaloneTagLine)
		}
		goto end
	}

	// Not eat the tag identifier and the closing braces
	multiline = false
	for *lineNo < len(tc.lines) {
		if pos >= len(line) {
			if !tagOpened {
				tc.AddLineType(*lineNo, TextLine)
				goto end
			}
			pos = 0
			*lineNo++
			line = tc.lines[*lineNo]
			multiline = true
			continue
		}
		if line[pos] != tc.closeDelim[0] {
			pos++
			continue
		}
		if line[pos+1] != tc.closeDelim[1] {
			err = fmt.Errorf("ERROR: closing delimiters not found for line '%s'", line)
			goto end
		}
		tagClosed = true
		if multiline {
			if notStandalone {
				tc.AddLineType(start, InlineTagLine)
			} else {
				tc.AddLineType(start, StandaloneTagLine)
			}
			tc.AddLineType(start, MultilineTagBegin)
			for i := start + 1; i < *lineNo; i++ {
				tc.AddLineType(i, MultilineTagMiddle)
			}
			tc.AddLineType(*lineNo, MultilineTagEnd)
		}
		pos += len(tc.closeDelim)
		break
	}

	// If we didn't find a closing delimiter, this isn't a valid tag
	if !tagClosed {
		tc.AddLineType(*lineNo, TextLine)
		goto end
	}
	if notStandalone {
		tc.AddLineType(*lineNo, InlineTagLine)
		goto end
	}
	// Now seek a second tag. If found, we have inline tags.
	// If not found, we have Standalone tags.
	tc.AddLineType(*lineNo, tc.seekSecondTag(line, &pos))

end:
	return err
}

func (tc *TemplateClassifier) tripleBracketFound(line string, pos *int, tb string) (found bool) {
	var start, n int
	if len(line) < 3 {
		goto end
	}
	start = *pos
	for *pos < len(line) && line[*pos] != tb[0] {
		*pos++
	}
	if *pos == len(line) {
		*pos = start
		goto end
	}
	n++
	for ; *pos < len(line); *pos++ {
		if line[*pos] != tb[n] {
			// We did not find a triple bracket begin, bail from this func
			goto end
		}
		n++
		if n == len(tb) {
			// We found a triple bracket
			*pos++
			found = line[*pos] == tb[n-1]
			goto end
		}
	}
end:
	return found
}

func (tc *TemplateClassifier) processTagPrefix(lineNo, pos *int) (tagClosed bool, line string, err error) {

	line = tc.lines[*lineNo]

	if *pos >= len(line) {
		*lineNo++
		*pos = 0
	}
	if *lineNo >= len(tc.lines) {
		goto end
	}
	line = tc.lines[*lineNo]
	tagClosed = true
	switch line[*pos] {
	case '!': // Comments
		err = tc.processComments(lineNo, pos)
	case '&': // Unescaped
		err = tc.processUnescaped(lineNo, pos)
	default:
		tagClosed = false
	}
end:
	return tagClosed, tc.lines[*lineNo], err
}

func (tc *TemplateClassifier) maybeEatTripleBracket(line string, pos *int) (ate bool, lt LineAspect, err error) {
	start := *pos
	if !tc.tripleBracketFound(line, pos, TripleBracketBegin) {
		// Not a triple bracket
		*pos = start
		goto end
	}
	if !tc.tripleBracketFound(line, pos, TripleBracketEnd) {
		err = fmt.Errorf(`parse error: did not find closing triple bracket in '%s'`, line)
		goto end
	}
	*pos++
	ate = true
	if *pos == len(line) {
		lt = StandaloneTagLine
		goto end
	}
	lt = tc.seekSecondTag(line, pos)
end:
	return ate, lt, err
}

func (tc *TemplateClassifier) seekSecondTag(line string, pos *int) (lt LineAspect) {
	lt = InlineTagLine

	// Check if there's another opening delimiter after the first tag
	lastPos := len(line) - len(tc.openDelim)

	tbb, od := TripleBracketBegin[0], tc.openDelim[0]

	start := *pos
	for ; *pos <= lastPos; *pos++ {
		switch line[*pos] {
		case tbb, od:
			break
		}
	}
	if *pos == lastPos {
		lt = StandaloneTagLine
		goto end
	}
	*pos = start
	if tc.tripleBracketFound(line, pos, TripleBracketBegin) {
		// Found a triple bracket delimiter - not a standalone tag
		goto end
	}
	for ; *pos <= lastPos; *pos++ {
		if tc.matchesDelim(line, *pos, tc.openDelim) {
			// Found another opening delimiter - not a standalone tag
			goto end
		}
	}

	// Check if there's only whitespace after the tag
	for *pos = start; *pos < len(line); *pos++ {
		if !tc.isNonNewlineWhitespace(line[*pos]) {
			// Found non-whitespace after the tag - not a standalone tag
			goto end
		}
	}
	lt = StandaloneTagLine
end:
	return lt
}

func (tc *TemplateClassifier) processUnescaped(lineNo *int, pos *int) (err error) {
	line := tc.lines[*lineNo]
	for ; *pos < len(line); *pos++ {
		if line[*pos] != tc.closeDelim[0] {
			continue
		}
		break
	}
	if *pos == len(line) {
		err = fmt.Errorf("ERROR: closing delimiters not found for line '%s'", line)
		goto end
	}
	tc.AddLineType(*lineNo, AmpersandUnescaped)
	*pos += len(tc.closeDelim)
end:
	return err
}

func (tc *TemplateClassifier) processComments(lineNo *int, pos *int) (err error) {
	tc.AddLineType(*lineNo, StandaloneTagLine)
	begin := *lineNo
	err = tc.eatComment(lineNo, pos)
	if err != nil {
		goto end
	}
	*pos += len(tc.closeDelim)
	if begin == *lineNo {
		goto end
	}
	tc.AddLineType(begin, MultilineTagBegin)
	for i := begin + 1; i < *lineNo; i++ {
		tc.AddLineType(i, MultilineTagMiddle)
	}
	tc.AddLineType(*lineNo, MultilineTagEnd)
end:
	return err
}

func (tc *TemplateClassifier) eatComment(index *int, pos *int) (err error) {
	for {
		line := tc.lines[*index]
		for ; *pos < len(line)-1; *pos++ {
			if line[*pos] != tc.closeDelim[0] {
				continue
			}
			if line[*pos+1] != tc.closeDelim[1] {
				continue
			}
			//*pos--
			goto end
		}
		*pos = 0
		*index++
	}
end:
	return err
}

func (tc *TemplateClassifier) isNonNewlineWhitespace(c byte) (is bool) {
	return c == ' ' || c == '\t' || c == '\r'
}

func (tc *TemplateClassifier) matchesDelim(line string, pos int, delim string) (matches bool) {
	if pos >= len(line)-1 {
		goto end
	}
	if line[pos] != delim[0] {
		goto end
	}
	if line[pos+1] != delim[1] {
		goto end
	}
	matches = true
end:
	return matches
}

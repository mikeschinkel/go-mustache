package classifier

import (
	"fmt"
	"regexp"

	"github.com/alexkappa/mustache"
)

const (
	TripleBraceBegin = "{{{"
	TripleBraceEnd   = "}}}"
)

type TemplateClassifier struct {
	template      string
	openDelim     string
	closeDelim    string
	templateLines []string
	lines         Lines
	tagStack      Stack[Tag]
}

// PastEOT returns true if the line number is past the end of template
func (tc *TemplateClassifier) PastEOT(lineNo *int) bool {
	return *lineNo >= len(tc.lines)
}

// AtEOT returns true if the line number is last line number of the template
func (tc *TemplateClassifier) AtEOT(lineNo *int) bool {
	return *lineNo == len(tc.lines)-1
}

var crlfRegex = regexp.MustCompile("(\r\n|\r|\n)")

func NewTemplateClassifier(template string) *TemplateClassifier {
	return &TemplateClassifier{
		template:   template,
		openDelim:  "{{",
		closeDelim: "}}",
	}
}

func (tc *TemplateClassifier) Initialize() (err error) {
	if tc.lines == nil {
		// Create a slice of Line
		contentLines := crlfRegex.Split(tc.template, -1)
		if contentLines == nil && len(tc.template) != 0 {
			tc.lines = Lines{{
				Segments: make(Segments, 0),
			}}
			tc.templateLines = []string{tc.template}
			goto end
		}
		tc.templateLines = make([]string, len(contentLines))
		tc.lines = make([]Line, len(contentLines))
		for i := range tc.lines {
			tc.lines[i] = Line{
				Segments: make(Segments, 0),
			}
			tc.templateLines[i] = contentLines[i]
		}
		goto end
	}
	// Reset the values that need resetting
	for i, line := range tc.lines {
		line.Segments = make(Segments, 0)
		line.Type = InvalidLineType
		tc.lines[i] = line
	}
	tc.tagStack = Stack[Tag]{}
end:
	return err
}

// Classify analyzes a mustache template string and returns
// a slice of booleans indicating whether each line contains a standalone tag.
// This implementation properly handles line boundaries and tag detection.
func (tc *TemplateClassifier) Classify() (_ Lines, err error) {
	errs := mustache.NewMultiErr()
	err = tc.Initialize()
	if err != nil {
		errs.Add(err)
		goto end
	}
	for i := 0; i < len(tc.lines); i++ {
		err = tc.classifyLine(&i)
		if err != nil {
			errs.Add(err)
		}
	}
end:
	if tc.tagStack.Depth() != 0 {
		errs.Add(fmt.Errorf("unclosed tags: %s", &tc.tagStack))
	}
	return tc.lines, errs.Err()
}

func (tc *TemplateClassifier) classifyLine(lineNo *int) (err error) {
	var ott OpenTagType

	pos := new(int)
	firstPass := true

	if tc.isEmptyLine(lineNo) {
		tc.SetLineType(lineNo, EmptyLine)
		goto end
	}

	for *pos < tc.lineLen(lineNo) {
		var ateWS bool
		var ateText bool

		// Skip leading whitespace
		ateWS = tc.maybeEatInlineWhitespace(lineNo, pos)
		// If during first pass we've reached the end of the line and it's just whitespace
		if firstPass {
			if tc.PastEOL(lineNo, pos) {
				tc.SetLineType(lineNo, WhitespaceLine)
				tc.AddSegment(lineNo, Segment{Type: Whitespace})
				goto end
			}
			if ateWS {
				tc.AddSegment(lineNo, Segment{Type: Whitespace})
			}
		} else if tc.PastEOL(lineNo, pos) && ateWS {
			tc.AddSegment(lineNo, Segment{Type: Whitespace})
		}

		// Eat any text and then classify the OpenTagTag found, if any
		ott, ateText = tc.maybeEatTextContent(lineNo, pos)
		if tc.PastEOL(lineNo, pos) {
			if ateText {
				tc.AddSegment(lineNo, Segment{Type: TextContent})
			}
			goto end
		}
		if !firstPass && !ateText && ateWS {
			tc.AddSegment(lineNo, Segment{Type: Whitespace})
		}
		if tc.PastEOT(lineNo) {
			goto end
		}

		switch ott {
		case DelimiterOpen:
			err = tc.maybeEatDelimitedTag(lineNo, pos)
		case TripleBraceOpen:
			err = tc.maybeEatTripleBrace(lineNo, pos)
		case InvalidOpenTag:
			fallthrough
		default:
			// This really should never happen as maybeEatTextContent() never returns an invalid
			// open tag unless it is past EOL or EOT in which case we never here.
			panic(fmt.Sprintf("Invalid Open Tag Type: %s", ott))
		}
		firstPass = false
	}
	if tc.LineType(lineNo) != InvalidLineType {
		goto end
	}
end:
	tc.SetLineType(lineNo, tc.getLineType(lineNo))
	return err
}

// getLineType returns the line type based upon the number of segments. the
// segment types, and the segment tag types, as applicable.
func (tc *TemplateClassifier) getLineType(lineNo *int) (lt LineType) {
	var segments Segments

	lt = tc.LineType(lineNo)
	if lt != InvalidLineType {
		goto end
	}
	segments = tc.LineSegments(lineNo)
	switch len(segments) {
	case 1:
		switch segments[0].Type {
		case TextContent:
			lt = TextLine
		case Whitespace:
			lt = WhitespaceLine
		case CompleteTag, BeginTag, EndTag:
			checkValidTagSegment(segments[0])
			lt = StandaloneLine
		default:
			lt = InlineLine
		}
	case 2:
		switch segments[0].Type {
		case Whitespace:
			if segments[1].Type == TextContent {
				lt = TextLine
				goto end
			}
			if segments[1].IsValidTagType() {
				lt = StandaloneLine
			}
		case CompleteTag:
			checkValidTagSegment(segments[0])
			if segments[1].Type == Whitespace {
				lt = StandaloneLine
			}
		default:
			lt = InlineLine
		}
	case 3:
		switch segments[0].Type {
		case Whitespace:
			if segments[1].IsValidTagType() && segments[2].Type == Whitespace {
				lt = StandaloneLine
			}
		default:
			lt = InlineLine
		}
	default:
		lt = InlineLine
	}
	if lt == InvalidLineType {
		lt = InlineLine
	}
end:
	return lt
}

// checkValidTagSegment verifies that a passed segment is indeed a valid tag
// segment. Used by getLineType().
func checkValidTagSegment(segment Segment) {
	if !segment.IsValidTagType() {
		panic(fmt.Sprintf("Unexpected mismatch between segment type and tag type: %+v", segment))
	}
}

// maybeEatDelimitedTag gets called right after we have validated an open
// delimiter was processed and pos should be pointing to the first character
// after the 2nd open delimiter at which point it should eat the tag up to and
// including its closing tag and/or brace. If the tag is an enclosing tag, it
// should recurse and eat all that it contains. It returns an error if the
// template is found to be malformed during parsing.
func (tc *TemplateClassifier) maybeEatDelimitedTag(lineNo, pos *int) (err error) {
	if tc.PastEOL(lineNo, pos) {
		// With the exception of comment tags — which get parsed separately from this
		// func — tags should never exist across newlines.
		err = fmt.Errorf("invalid tag format; open delimeters ('%s') without closing delimieters *'%s') for line %s",
			tc.openDelim,
			tc.closeDelim,
			tc.Template(lineNo),
		)
		goto end
	}
	if tc.PastEOT(lineNo) {
		// With the exception of comment tags — which get parsed separately from this
		// func — tags should never exist across newlines.
		err = fmt.Errorf("invalid tag format; open delimeters ('%s') without additional template content for line %s",
			tc.openDelim,
			tc.Template(lineNo),
		)
		goto end
	}
	switch tc.char(lineNo, pos) {
	case '.': // Dot
		err = tc.parseDots(lineNo, pos)
	case '>': // Partial
		err = tc.parsePartials(lineNo, pos)
	case '=': // Set Delimiter
		err = tc.parseSetDelimiters(lineNo, pos)
	case '!': // Comments
		err = tc.parseComments(lineNo, pos)
	case '&': // Unescaped
		err = tc.parseUnescaped(lineNo, pos)
	case '#': // Section
		err = tc.parseSections(lineNo, pos)
	case '^': // Inverted Section
		err = tc.parseInvertedSections(lineNo, pos)
	case '<': // Parent
		err = tc.parseParents(lineNo, pos)
	case '$': // Block
		err = tc.parseBlocks(lineNo, pos)
	case '/': // Closing
		err = tc.parseClosings(lineNo, pos)
	default:
		err = tc.parseVars(lineNo, pos)
	}
end:
	return err
}

func (tc *TemplateClassifier) maybeEatInlineWhitespace(lineNo, pos *int) (ateSome bool) {
	var begin = *pos
	// Advanced through inline whitespace
	for !tc.AtEOL(lineNo, pos) {
		if !isInlineWhitespace(tc.char(lineNo, pos)) {
			break
		}
		*pos++
	}
	return *pos > begin
}

// maybeEatText eats text, if any, until it either reaches an open tag or the end
// of the line.
func (tc *TemplateClassifier) maybeEatTextContent(lineNo *int, pos *int) (ott OpenTagType, ateSome bool) {
	begin := *pos
	for !tc.PastEOL(lineNo, pos) {
		switch tc.char(lineNo, pos) {
		case TripleBraceBegin[0], tc.openDelim[0]:
			// We've come to the end of text
			// Classify the open tag type
			ott = tc.getOpenTagType(lineNo, pos)
			if ott == InvalidOpenTag {
				// The first char of an open tag was just text. Keep on maybe eating.
				continue
			}
			if ott == TripleBraceOpen {
				tc.AddSegment(lineNo, Segment{TagType: TripleBraceUnescaped})
				*pos += 2
				goto end
			}
			if *pos > begin {
				// We did eat text, so note the fact
				tc.AddSegment(lineNo, Segment{Type: TextContent})
			}
			*pos += 2
			goto end
		default:
			// Not a special char, so advance to next position and keep maybe eating text
			*pos++
		}
	}
	ateSome = *pos > begin
end:
	return ott, ateSome
}

// maybeEatTagClosing eats the closing delimiters, e.g. '}}' or whatever has been
// set as custom delimiters and leaves *pos pointing to the character immediately
// after, i.e. for `{{foo}}bar` *pos would be 7 and its character would be 'b'.
func (tc *TemplateClassifier) maybeEatTagClosingDelimiters(lineNo *int, pos *int, tagType OpenTagType) (err error) {
	var name string
	t := tc.templateLines[*lineNo]
	switch tagType {
	case DelimiterOpen:
		if len(t)-*pos >= 2 && t[*pos:*pos+2] == tc.closeDelim {
			*pos += 2
			goto end
		}
		// Not a valid closing delimiter
		name = "delimiter"
	case TripleBraceOpen:
		if len(t)-*pos >= 3 && t[*pos:*pos+3] == TripleBraceEnd {
			*pos += 3
			goto end
		}
		// Not a valid closing triple brace
		name = "triple-brace"
	case InvalidOpenTag:
		// This is an error so just run the next line
	}
	err = fmt.Errorf("invalid closing %s tag in '%s'", name, tc.Template(lineNo))
end:
	return err
}

func (tc *TemplateClassifier) maybeEatIdentifier(lineNo *int, pos *int) (identifier string, err error) {
	begin := *pos
	for !tc.AtEOL(lineNo, pos) {
		switch tc.char(lineNo, pos) {
		case TripleBraceEnd[0], tc.closeDelim[0]:
			// We've come to the end of identifier
			if *pos == begin {
				err = fmt.Errorf("tag identifier is empty in line %s", tc.Template(lineNo))
				goto end
			}
			identifier = tc.TemplateSubstring(lineNo, begin, *pos)
			goto end
		default:
			// Not a special char, so advance to next position and keep maybe eating text
			*pos++
		}
	}
end:
	return identifier, err
}

func (tc *TemplateClassifier) getOpenTagType(lineNo, pos *int) (ott OpenTagType) {
	t := tc.templateLines[*lineNo]
	remaining := len(t) - *pos

	if remaining < 2 {
		// If less than 2 characters left, this can't be an open delimiter
		goto end
	}
	if t[*pos] != TripleBraceBegin[0] {
		// If first char is not equal to the first known char of triple brace it cannot
		// be a triple brace.
	}
	if t[*pos+1] != TripleBraceBegin[1] {
		// If next char is not equal to the second known char of triple brace it cannot
		// be a triple brace.
	}
	if remaining < 3 || t[*pos+2] != TripleBraceBegin[2] {
		// If less than 3 characters left, can't be an open triple brace
		if t[*pos] != tc.openDelim[0] {
			// Cannot be an open tag if these don't match
			goto end
		}
		if t[*pos+1] != tc.openDelim[1] {
			// Cannot be an open tag if these don't match
			goto end
		}
		ott = DelimiterOpen
		goto end
	}
	// We checked all three and it is a triple brace
	ott = TripleBraceOpen
end:
	return ott
}

// maybeEatTripleBrace gets called right after we have validated an open triple
// brace pos should be pointing to the first character after the 3rd open triple
// brace at which point it should eat the tag the identifier and then its closing
// triple brace. It returns an error if the template is found to be malformed
// during parsing.
func (tc *TemplateClassifier) maybeEatTripleBrace(lineNo, pos *int) (err error) {
	_, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, TripleBraceOpen)
	if err != nil {
		goto end
	}
end:
	return err
}

func (tc *TemplateClassifier) parseUnescaped(lineNo *int, pos *int) (err error) {
	tl := tc.templateLines[*lineNo]
	for !tc.PastEOL(lineNo, pos) {
		*pos++
		if tl[*pos] != tc.closeDelim[0] {
			continue
		}
		break
	}
	if tc.PastEOL(lineNo, pos) {
		err = fmt.Errorf("ERROR: closing delimiters not found for line '%s'", tl)
		goto end
	}
	tc.AddSegment(lineNo, Segment{TagType: AmpersandUnescaped})
	*pos += len(tc.closeDelim)
end:
	return err
}

func (tc *TemplateClassifier) parseComments(lineNo *int, pos *int) (err error) {
	var ateFull bool
	var segments Segments
	var lt LineType

	firstLine := true
	begin := *lineNo
	*pos++ // Eat '!' comment char
	for {
		ateFull, err = tc.eatFullComment(lineNo, pos, firstLine, begin < *lineNo)
		if err != nil {
			goto end
		}
		if ateFull {
			break
		}
		if tc.pastEnd(lineNo, pos) == NotAtEnd {
			firstLine = false
			continue
		}
		break
	}
	segments = tc.LineSegments(lineNo)
	if *lineNo > begin {
		// Multiline
		lt = InlineLine
		goto end
	}
	lt = StandaloneLine
	// Single line
	if len(segments) == 1 {
		// Only a comment
		goto end
	}
	if len(segments) == 2 && segments[0].Type == Whitespace {
		// A comment with only leading whitespace
		goto end
	}
	if len(segments) == 3 && segments[0].Type == Whitespace && segments[2].Type == Whitespace {
		// A comment with only leading and trailing whitespace
		goto end
	}
	lt = InlineLine
end:
	for begin <= *lineNo {
		tc.SetLineType(&begin, lt)
		begin++
	}
	return err
}

type EndType int

const (
	NotAtEnd EndType = iota
	EOL
	EOT
)

// pastEnd returns true if either pos is past end of line, or if lineNo is past
// end of template. If past EOL it sets pos to zero and also increments lineNo
// the latter of which (BTW) can trigger past EOT. Returns one of EOL, EOT, or
// NotAtEnd.
func (tc *TemplateClassifier) pastEnd(lineNo *int, pos *int) (et EndType) {
	if tc.PastEOL(lineNo, pos) {
		*lineNo++
		*pos = 0
		et = EOL
	}
	if tc.PastEOT(lineNo) {
		et = EOT
	}
	return et
}

func (tc *TemplateClassifier) addCommentSegment(lineNo int, multiline bool, multilineType SegmentType, firstLine bool, nonFirstLineType SegmentType) {
	st := CompleteTag
	if multiline {
		st = multilineType
	}
	segment := Segment{Type: st, TagType: CommentTag}
	if !firstLine && multiline {
		segment.Type = nonFirstLineType
	}
	tc.AddSegment(&lineNo, segment)
}

func (tc *TemplateClassifier) eatFullComment(lineNo *int, pos *int, firstLine bool, multiline bool) (ateClose bool, err error) {

	if tc.char(lineNo, pos) == tc.closeDelim[0] {
		// Closing delimiter
		*pos++
		et := tc.pastEnd(lineNo, pos)
		switch et {
		case EOL:
			multiline = true
			tc.addCommentSegment(*lineNo-1, multiline, MultilineBegin, firstLine, MultilineMiddle)
		case EOT:
			err = fmt.Errorf("invalid comment tag in template line '%s'", tc.line(lineNo))
			goto end
		case NotAtEnd:
			if tc.char(lineNo, pos) == tc.closeDelim[1] {
				tc.addCommentSegment(*lineNo, multiline, MultilineMiddle, firstLine, MultilineEnd)
				*pos++
				ateClose = true
				goto end
			}
		}
		// ch != 2nd closing delimiter so we were looking at text content after all.
		*pos--
	}
	// Text content (includes non-leading whitespace)
	for {
		*pos++
		et := tc.pastEnd(lineNo, pos)
		switch et {
		case EOL:
			multiline = true
			tc.addCommentSegment(*lineNo-1, multiline, MultilineBegin, firstLine, MultilineMiddle)
		case EOT:
			err = fmt.Errorf("invalid comment tag in template '%s'", tc.template)
			goto end
		case NotAtEnd:
			if tc.char(lineNo, pos) != tc.closeDelim[0] {
				continue
			}
			goto end
		}
		goto end
	}
end:
	return ateClose, err
}

// Process section tags (#)
func (tc *TemplateClassifier) parseSections(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, SectionTag)
}

// Process inverted section tags (^)
func (tc *TemplateClassifier) parseInvertedSections(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, InvertedSectionTag)
}

// Process closing tags (/)
func (tc *TemplateClassifier) parseClosings(lineNo *int, pos *int) (err error) {
	var identifier string
	var tag Tag

	*pos++ // Eat closing char '/'

	begin := *pos

	// Consume the section name until the closing delimiter
	identifier, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}

	tag = tc.tagStack.Pop()
	if tag.Identifier != identifier {
		template := tc.Template(lineNo)
		err = fmt.Errorf("ERROR: mismatched open and close delimiters for line '%s';open='%s', close='%s'",
			template,
			tag.Identifier,
			template[begin:*pos],
		)
	}

	tc.AddSegment(lineNo, Segment{
		TagType: tag.Type,
		Type:    EndTag,
	})

end:
	return err
}

// Process partial tags (>)
func (tc *TemplateClassifier) parseVars(lineNo *int, pos *int) (err error) {
	var identifier string

	identifier, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}

	switch identifier {
	case ".":
		tc.AddSegment(lineNo, Segment{TagType: DotTag})
	default:
		tc.AddSegment(lineNo, Segment{TagType: VarTag})
	}

end:
	return err
}

// Process partial tags (>)
func (tc *TemplateClassifier) parsePartials(lineNo *int, pos *int) (err error) {
	*pos++ // Eat enclosing partial identifying character '>'

	// Eat partial identifier
	_, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	// Eat tag closing
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}
	tc.AddSegment(lineNo, Segment{TagType: PartialTag})
end:
	return err
}

// Process parent tags (<)
func (tc *TemplateClassifier) parseParents(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, ParentTag)
}

// Process block tags ($)
func (tc *TemplateClassifier) parseBlocks(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, BlockTag)
}

type numeric interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

// Process set delimiter tags (=)
func (tc *TemplateClassifier) parseSetDelimiters(lineNo *int, pos *int) (err error) {
	var newOpen, newClose string
	var et EndType

	for {
		endOfDelimiter := addToPtr(pos, 2)
		if tc.PastEOL(lineNo, endOfDelimiter) {
			err = fmt.Errorf("invalid delimiters parsing custom delimiters in line '%s'", tc.line(lineNo))
			goto end
		}
		et = tc.pastEnd(lineNo, pos)
		switch et {
		case EOL:
			err = fmt.Errorf("unexpected end of line (EOL) parsing custom delimiters in line '%s'", tc.line(lineNo))
			goto end
		case EOT:
			err = fmt.Errorf("unexpected end of template (EOT) parsing custom delimiters in line '%s'", tc.line(lineNo))
			goto end
		case NotAtEnd:
			if tc.char(lineNo, pos) != '=' {
				*pos++
				continue
			}
			*pos++
			if newOpen == "" {
				// Extract the new closing template
				newOpen = tc.TemplateSubstring(lineNo, *pos, *pos+2)
				*pos += 2
				continue
			}
			// Extract the new closing template
			newClose = tc.TemplateSubstring(lineNo, *pos-3, *pos-1)
			err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
			tc.openDelim = newOpen
			tc.closeDelim = newClose
			tc.AddSegment(lineNo, Segment{Type: CompleteTag, TagType: DelimiterTag})
			goto end
		}
	}
end:
	return err
}

// Process dot tags (.)
func (tc *TemplateClassifier) parseDots(lineNo *int, pos *int) (err error) {
	*pos++ // Eat dot '.'

	// Eat tag closing
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}
	tc.AddSegment(lineNo, Segment{TagType: DotTag})
end:
	return err
}

// parseEnclosingBeginTag processes the opening tag for all enclosing tags, e.g. section, block, etc.
func (tc *TemplateClassifier) parseEnclosingBeginTag(lineNo *int, pos *int, tagType TagType) (err error) {
	var identifier string

	*pos++ // Eat enclosing tag identifying character, e.g. '#', '$', '^', etc.

	// Eat identifier for parent tag
	identifier, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	// Eat tag closing delimiters
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}
	tc.AddSegment(lineNo, Segment{TagType: tagType, Type: BeginTag})
	tc.tagStack.Push(Tag{
		Type:       tagType,
		Identifier: identifier,
	})
end:
	return err
}

func (tc *TemplateClassifier) SetLineType(lineNo *int, lt LineType) {
	tc.ensureLines(lineNo)
	tc.lines[*lineNo].Type = lt
}
func (tc *TemplateClassifier) AddSegment(lineNo *int, segment Segment) {
	tc.ensureLines(lineNo)
	tc.lines[*lineNo].Segments = append(tc.lines[*lineNo].Segments, segment)
}
func (tc *TemplateClassifier) ensureLines(lineNo *int) {
	nextLineNo := *lineNo + 1
	if nextLineNo > len(tc.lines) {
		tc.templateLines = make([]string, nextLineNo)
		tc.lines = make(Lines, nextLineNo)
	}
}
func (tc *TemplateClassifier) line(lineNo *int) *Line {
	return &tc.lines[*lineNo]
}
func (tc *TemplateClassifier) Template(lineNo *int) string {
	return tc.templateLines[*lineNo]
}
func (tc *TemplateClassifier) LineType(lineNo *int) LineType {
	return tc.lines[*lineNo].Type
}
func (tc *TemplateClassifier) TemplateSubstring(lineNo *int, begin, end int) string {
	return tc.templateLines[*lineNo][begin:end]
}
func (tc *TemplateClassifier) char(lineNo, pos *int) byte {
	return tc.templateLines[*lineNo][*pos]
}
func (tc *TemplateClassifier) lineLen(lineNo *int) int {
	return len(tc.templateLines[*lineNo])
}
func (tc *TemplateClassifier) isEmptyLine(lineNo *int) bool {
	return len(tc.templateLines[*lineNo]) == 0
}
func (tc *TemplateClassifier) AtEOL(lineNo, pos *int) bool {
	return tc.lines[*lineNo].AtEOL(pos, tc.lineLen(lineNo))
}
func (tc *TemplateClassifier) PastEOL(lineNo, pos *int) bool {
	return tc.lines[*lineNo].PastEOL(pos, tc.lineLen(lineNo))
}
func (tc *TemplateClassifier) LineSegments(lineNo *int) (ss Segments) {
	return tc.lines[*lineNo].Segments
}

package classifier

import (
	"errors"
	"regexp"

	"github.com/alexkappa/mustache"
)

// Constants for triple brace delimiters used for unescaped variables.
const (
	// TripleBraceBegin represents the opening triple brace delimiter "{{{".
	TripleBraceBegin = "{{{"

	// TripleBraceEnd represents the closing triple brace delimiter "}}}".
	TripleBraceEnd = "}}}"
)

// TemplateClassifier is responsible for analyzing Mustache templates
// and classifying each line to determine how it should be rendered.
// It tracks template state, line content, and tag nesting.
type TemplateClassifier struct {
	// template is the raw template string.
	template string

	// openDelim is the current opening delimiter (default: "{{").
	openDelim string

	// closeDelim is the current closing delimiter (default: "}}").
	closeDelim string

	// templateLines contains the template lines split by newlines ('\n')
	// and/or CRLFs ("\r\n")
	templateLines []string

	// lines contains the classification of the templateLines which includes
	// a type for each templateLine and one of more Segments for each line
	// where a Segment contains SegmentType and a TagType.
	lines Lines

	// tagStack tracks nested tags (sections, inverted sections, etc.).
	tagStack Stack[Tag]
}

// PastEOT returns true if the line number is past the end of template.
func (tc *TemplateClassifier) PastEOT(lineNo *int) bool {
	return *lineNo >= len(tc.lines)
}

// AtEOT returns true if the line number is the last line number of the template.
func (tc *TemplateClassifier) AtEOT(lineNo *int) bool {
	return *lineNo == len(tc.lines)-1
}

// crlfRegex is a regular expression that matches all possible line endings (CR,
// LF, CRLF).
var crlfRegex = regexp.MustCompile("(\r\n|\r|\n)")

// NewTemplateClassifier creates a new TemplateClassifier with the provided
// template string and it initializes the default delimiters "{{" and "}}".
func NewTemplateClassifier(template string) *TemplateClassifier {
	return &TemplateClassifier{
		template:   template,
		openDelim:  "{{",
		closeDelim: "}}",
	}
}

// Initialize sets up the classifier's internal state for parsing. It splits the
// template into lines and initializes all required data structures. If already
// initialized, it resets the classifier to process the template again.
func (tc *TemplateClassifier) Initialize() (err error) {
	if tc.lines == nil {
		// Create a slice of Line
		contentLines := crlfRegex.Split(tc.template, -1)
		if contentLines == nil && len(tc.template) != 0 {
			tc.lines = Lines{{
				Segments: NewSegments(),
			}}
			tc.templateLines = []string{tc.template}
			goto end
		}
		tc.templateLines = make([]string, len(contentLines))
		tc.lines = make([]Line, len(contentLines))
		for i := range tc.lines {
			tc.lines[i] = Line{
				Segments: NewSegments(),
			}
			tc.templateLines[i] = contentLines[i]
		}
		goto end
	}
	// Reset the values that need resetting
	for i, line := range tc.lines {
		line.Segments = NewSegments()
		line.Type = InvalidLineType
		tc.lines[i] = line
	}
	tc.tagStack = Stack[Tag]{}
end:
	return err
}

// Classify analyzes a Mustache template string and classifies each line. It
// returns a slice of Line objects with their classifications and segments. The
// classification determines how each line should be rendered. It also checks for
// proper nesting of tags and returns errors for unclosed tags.
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
		errs.Add(errors.Join(ErrUnclosedTags,
			ErrArg(TagStackErrArg, &tc.tagStack),
		))
	}
	return tc.lines, errs.Err()
}

// classifyLine processes a single line of the template to determine its
// classification and identify all segments within it. It handles whitespace,
// text content, and various tag types.
func (tc *TemplateClassifier) classifyLine(lineNo *int) (err error) {
	var ott OpenTagType

	pos := new(int)
	firstPass := true

	// Handle empty lines quickly
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
				tc.AddSegment(lineNo, MakeWhitespaceSegment())
				goto end
			}
			if ateWS {
				tc.AddSegment(lineNo, MakeWhitespaceSegment())
			}
		} else if tc.PastEOL(lineNo, pos) && ateWS {
			tc.AddSegment(lineNo, MakeWhitespaceSegment())
		}

		// Eat any text and then classify the OpenTagTag found, if any
		ott, ateText = tc.maybeEatTextContent(lineNo, pos)
		if tc.PastEOL(lineNo, pos) {
			goto end
		}
		if !firstPass && !ateText && ateWS {
			tc.AddSegment(lineNo, MakeWhitespaceSegment())
		}
		if tc.PastEOT(lineNo) {
			goto end
		}

		// Process the appropriate tag type
		switch ott {
		case DelimiterOpen:
			err = tc.maybeEatDelimitedTag(lineNo, pos)
		case TripleBraceOpen:
			err = tc.maybeEatTripleBrace(lineNo, pos)
		case InvalidOpenTag:
			// This really should never happen as maybeEatTextContent() never returns an
			// invalid open tag unless it is past EOL or EOT in which case we will never
			// get here.
			//goland:noinspection GoDfaNilDereference
			err = errors.Join(ErrInvalidTagOpeningType,
				ErrArg(TagOpenTypeErrArg, ott.String()),
				ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
			)
		}
		firstPass = false
	}
	if tc.LineType(lineNo) != InvalidLineType {
		goto end
	}
end:
	err = tc.maybeSetLineType(err, lineNo)
	return err
}

// maybeSetLineType sets the LineType for a line when no error has occurred. It
// determines the line type based on the segments found in the line. If an error
// is passed in or occurs during type determination, no line type is set.
func (tc *TemplateClassifier) maybeSetLineType(err error, lineNo *int) error {
	var lt LineType
	if err != nil {
		goto end
	}
	lt, err = tc.getLineType(lineNo)
	if err != nil {
		goto end
	}
	tc.SetLineType(lineNo, lt)
end:
	return err
}

// getLineType determines the line type based upon the number of segments, the
// segment types, and the segment tag types in the line. It implements the logic
// for distinguishing between different line types according to the Mustache
// specification.
func (tc *TemplateClassifier) getLineType(lineNo *int) (lt LineType, err error) {
	var segments Segments

	lt = tc.LineType(lineNo)
	if lt != InvalidLineType {
		goto end
	}
	segments = tc.LineSegments(lineNo)
	switch len(segments) {
	case 1:
		switch segments[0].Type() {
		case TextContent:
			lt = TextLine
		case Whitespace:
			lt = WhitespaceLine
		case CompleteTag, BeginTag, EndTag:
			err = checkValidTagSegment(segments[0])
			if err != nil {
				goto end
			}
			lt = StandaloneLine
		default:
			lt = InlineLine
		}
	case 2:
		switch segments[0].Type() {
		case Whitespace:
			if segments[1].Type() == TextContent {
				lt = TextLine
				goto end
			}
			if segments[1].IsValidTagType() {
				lt = StandaloneLine
			}
		case CompleteTag:
			err = checkValidTagSegment(segments[0])
			if err != nil {
				goto end
			}
			if segments[1].Type() == Whitespace {
				lt = StandaloneLine
			}
		default:
			lt = InlineLine
		}
	case 3:
		switch segments[0].Type() {
		case Whitespace:
			if segments[1].IsValidTagType() && segments[2].Type() == Whitespace {
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
	return lt, err
}

// checkValidTagSegment verifies that a segment has a valid tag type. This is
// used by getLineType() to validate tag segments before classifying a line.
// Returns an error if the segment's tag type is not valid.
func checkValidTagSegment(segment Segment) (err error) {
	if segment.IsValidTagType() {
		goto end
	}
	err = errors.Join(ErrSegmentTypeMismatch,
		ErrArg(SegmentErrArg, segment),
	)
end:
	return err
}

// maybeEatDelimitedTag processes a tag after an opening delimiter has been
// found. It identifies the tag type based on the first character and delegates
// to the appropriate parsing method. Position should be pointing to the first
// character after the opening delimiter.
//
// With the exception of comment tags (which are handled separately), tags should
// never span across newlines. This method returns an error if the template is
// malformed during parsing.
func (tc *TemplateClassifier) maybeEatDelimitedTag(lineNo, pos *int) (err error) {
	if tc.PastEOL(lineNo, pos) {
		// With the exception of comment tags — which get parsed separately from this
		// func — tags should never exist across newlines.
		err = errors.Join(ErrInvalidTagFormat,
			ErrArg(OpenDelimErrArg, tc.openDelim),
			ErrArg(CloseDelimErrArg, tc.closeDelim),
			ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
		)
		goto end
	}
	if tc.PastEOT(lineNo) {
		// With the exception of comment tags — which get parsed separately from this
		// func — tags should never exist across newlines.
		err = errors.Join(ErrInvalidTagFormat,
			ErrArg(OpenDelimErrArg, tc.openDelim),
			ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
		)
		goto end
	}

	// Dispatch to the appropriate tag parser based on the first character
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
	default: // Variable
		err = tc.parseVars(lineNo, pos)
	}
end:
	return err
}

// maybeEatInlineWhitespace advances the position pointer through any inline
// whitespace. Returns true if it consumed any whitespace, false otherwise.
func (tc *TemplateClassifier) maybeEatInlineWhitespace(lineNo, pos *int) (ateSome bool) {
	var begin = *pos
	// Advance through inline whitespace
	for !tc.AtEOL(lineNo, pos) {
		if !isInlineWhitespace(tc.char(lineNo, pos)) {
			break
		}
		*pos++
	}
	return *pos > begin
}

// maybeEatTextContent consumes regular text content until it finds an opening
// tag or reaches the end of the line. It identifies the type of opening tag
// found (if any) and updates the position pointer accordingly.
//
// Returns:
//   - ott:     Type of opening tag: DelimiterOpen, TripleBraceOpen, or
//     InvalidOpenTag.
//   - ateText: true if any text content was consumed, false otherwise
func (tc *TemplateClassifier) maybeEatTextContent(lineNo *int, pos *int) (ott OpenTagType, ateText bool) {
	openDelimLen := len(tc.openDelim)
	begin := *pos
	for !tc.PastEOL(lineNo, pos) {
		switch tc.char(lineNo, pos) {
		case TripleBraceBegin[0], tc.openDelim[0]:
			// We've come to the end of text
			// Classify the open tag type
			ott = tc.getOpenTagType(lineNo, pos)
			switch ott {
			case TripleBraceOpen:
				*pos += 2
				tc.AddSegment(lineNo, MakeTagSegment(TripleBraceUnescaped))
				goto end
			case DelimiterOpen:
				ateText = *pos > begin
				*pos += openDelimLen
				goto end
			case InvalidOpenTag:
				// The first char of an open tag was just text. Keep on maybe eating.
				*pos++
				continue
			}
		default:
			// Not a special char, so advance to next position and keep maybe eating text
			*pos++
			if tc.AtEOL(lineNo, pos) {
				// Well, we ate text content until EOL, so record that we did and leave
				ateText = true
				goto end
			}
		}
	}
end:
	if ateText {
		tc.AddSegment(lineNo, MakeTextContentSegment())
	}
	return ott, ateText
}

// maybeEatTagClosingDelimiters consumes the closing delimiters for a tag. It
// handles both standard delimiters (default: '}}') and triple braces ('}}}').
// After successful execution, the position pointer will point to the character
// immediately after the closing delimiter.
//
// For example, with `{{foo}}bar`, after calling this method with position at
// 'f', the position would point to 'b' (index 7).
func (tc *TemplateClassifier) maybeEatTagClosingDelimiters(lineNo *int, pos *int, tagType OpenTagType) (err error) {
	var errArg string
	t := tc.templateLines[*lineNo]
	switch tagType {
	case DelimiterOpen:
		delimLen := len(tc.closeDelim)
		if *pos+delimLen <= len(t) && t[*pos:*pos+delimLen] == tc.closeDelim {
			*pos += delimLen
			goto end
		}
		// Not a valid closing delimiter
		errArg = StandardDelimiterErrArg
	case TripleBraceOpen:
		if *pos+3 <= len(t) && t[*pos:*pos+3] == TripleBraceEnd {
			*pos += 3
			goto end
		}
		// Not a valid closing triple brace
		errArg = TripleBraceErrArg
	case InvalidOpenTag:
		// This is an error so just run the next line
	}
	err = errors.Join(ErrInvalidClosingTag,
		ErrArg(DelimiterTypeErrArg, errArg),
		ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
	)
end:
	return err
}

// maybeEatIdentifier consumes a tag identifier until it finds a closing
// delimiter or reaches the end of the line.
//
// For example, in `{{foobar}}`, this would consume "foobar".
// Returns the identifier string and any error that occurred during parsing.
func (tc *TemplateClassifier) maybeEatIdentifier(lineNo *int, pos *int) (identifier string, err error) {
	begin := *pos
	for !tc.AtEOL(lineNo, pos) {
		switch tc.char(lineNo, pos) {
		case TripleBraceEnd[0], tc.closeDelim[0]:
			// We've come to the end of identifier
			if *pos == begin {
				err = errors.Join(ErrMissingTagIdentifier,
					ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
				)
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

// maybeEatTripleBraceOpen checks if the current position contains a triple brace
// opening sequence '{{{' Returns true if it found a triple brace, false
// otherwise.
func (tc *TemplateClassifier) maybeEatTripleBraceOpen(lineNo, pos *int) (ate bool) {
	tl := tc.templateLines[*lineNo]
	if *pos+len(TripleBraceBegin) >= len(tl) {
		goto end
	}
	ate = true
	for i, c := range TripleBraceBegin {
		if byte(c) != tl[*pos+i] {
			ate = false
			goto end
		}
	}
end:
	return ate
}

// getOpenTagType determines what type of opening tag is at the current position.
// It can identify triple braces '{{{' or standard delimiters (default: '{{')
// Returns the identified OpenTagType (TripleBraceOpen, DelimiterOpen, or
// InvalidOpenTag)
func (tc *TemplateClassifier) getOpenTagType(lineNo, pos *int) (ott OpenTagType) {
	openDelimLen := len(tc.openDelim)
	//begin := *pos
	tl := tc.templateLines[*lineNo]
	remaining := len(tl) - *pos

	if remaining < openDelimLen {
		// If less than # open delimiter characters left, this can't be an open delimiter
		goto end
	}
	if tc.maybeEatTripleBraceOpen(lineNo, pos) {
		ott = TripleBraceOpen
		goto end
	}
	// If less than 3 characters left, can't be an open triple brace
	for delimPos := 0; delimPos < len(tc.openDelim); delimPos++ {
		if remaining < 0 {
			goto end
		}
		if tl[*pos+delimPos] != tc.openDelim[delimPos] {
			// Not a delimiter. Okay, just return
			goto end
		}
		remaining--
	}
	ott = DelimiterOpen
	goto end
end:
	return ott
}

// maybeEatTripleBrace processes a triple brace tag after identifying its opening
// sequence. It consumes the identifier and closing triple brace. Position should
// point to the first character after the triple brace opening sequence '{{{'.
//
// For example, in "{{{foobar}}}", it would consume "foobar}}}". Returns an error
// if the template is malformed during parsing.
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

// parseUnescaped processes an unescaped variable tag that uses the ampersand
// syntax {{&var}}. It consumes the content after the '&' character and adds an
// AmpersandUnescaped segment. Position should point to the '&' character at the
// beginning of the call. Returns an error if the tag is malformed.
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
		err = errors.Join(ErrInvalidTagFormat,
			ErrArg(TemplateLineErrArg, tl),
		)
		goto end
	}
	tc.AddSegment(lineNo, MakeTagSegment(AmpersandUnescaped))
	*pos += len(tc.closeDelim)
end:
	return err
}

// parseComments processes a comment tag {{! comment }}. Comments can span
// multiple lines. This method handles both single-line and multi-line comments.
// Position should point to the '!' character at the beginning of the call.
// Returns an error if the comment tag is malformed.
func (tc *TemplateClassifier) parseComments(lineNo *int, pos *int) (err error) {
	var ateFull bool
	var segments Segments
	var lt LineType
	var s0t SegmentType

	firstLine := true
	begin := *lineNo
	*pos++ // Eat '!' comment char

	// Process the comment content, potentially spanning multiple lines
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

	// Determine the line type based on the content and structure
	segments = tc.LineSegments(lineNo)
	if *lineNo > begin {
		// Multiline comment - always inline
		lt = InlineLine
		goto end
	}

	// Single line comment - could be standalone or inline
	lt = StandaloneLine

	if len(segments) == 1 {
		// Only a comment - standalone
		goto end
	}
	s0t = segments[0].Type()
	if len(segments) == 2 && s0t == Whitespace {
		// A comment with only leading whitespace - standalone
		goto end
	}
	if len(segments) == 3 && s0t == Whitespace && segments[2].Type() == Whitespace {
		// A comment with only leading and trailing whitespace - standalone
		goto end
	}

	// Any other pattern - inline
	lt = InlineLine

end:
	// Set the line type for all lines that were part of this comment
	for begin <= *lineNo {
		tc.SetLineType(&begin, lt)
		begin++
	}
	return err
}

// EndType represents different end-of-content conditions during parsing.
type EndType int

const (
	// NotAtEnd indicates the parser has not reached any end condition.
	NotAtEnd EndType = iota

	// EOL indicates the parser has reached the end of the current line.
	EOL

	// EOT indicates the parser has reached the end of the template.
	EOT
)

// pastEnd checks if the current position is past the end of the line or
// template. If past EOL, it increments the line number and resets position to 0.
// This method ensures line transitions are handled correctly without requiring
// the repeated occurrence repeating boilerplate logic.
//
// Returns:
// - EOL if past end of line (and advances to next line)
// - EOT if past end of template
// - NotAtEnd if not past any boundary
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

// notPastEnd provides error handling for end conditions during parsing.
// It checks if the current position is past any boundary and generates
// appropriate error messages. If not past any boundary, it executes the
// provided function.
//
// Parameters:
// - lineNo: Pointer to the current line number
// - pos: Pointer to the current position
// - ifNot: Function to call if not past any boundary
//
// Returns an appropriate error based on the end condition.
func (tc *TemplateClassifier) notPastEnd(lineNo *int, pos *int, ifNot func() error) (err error) {
	et := tc.pastEnd(lineNo, pos)
	switch et {
	case EOL:
		err = errors.Join(ErrUnexpectedEOL,
			ErrArg(TemplateLineErrArg, tc.line(lineNo)),
		)
	case EOT:
		err = errors.Join(ErrUnexpectedEOT,
			ErrArg(TemplateLineErrArg, tc.line(lineNo)),
		)
	default:
		err = ifNot()
	}
	return err
}

// addCommentSegment adds the appropriate comment segment to a line based on its
// position within a multi-line comment (if applicable).
//
// Parameters:
//   - lineNo:           Line number to add the segment to
//   - multiline:        Whether this is part of a multi-line comment
//   - multilineType:    Segment type to use for first lines in multiline comments
//   - firstLine:        Whether this is the first line of a comment
//   - nonFirstLineType: The segment type to use for non-first lines in multiline
//     comments
func (tc *TemplateClassifier) addCommentSegment(lineNo int, multiline bool, multilineType SegmentType, firstLine bool, nonFirstLineType SegmentType) {
	st := CompleteTag
	if multiline {
		st = multilineType
	}
	segment := MakeCommentTagSegment(st)
	if !firstLine && multiline {
		segment = MakeCommentTagSegment(nonFirstLineType)
	}
	tc.AddSegment(&lineNo, segment)
}

// eatFullComment processes the content of a comment tag and handles multi-line
// comments. It consumes text until it finds the closing delimiter or reaches the
// end of the template.
//
// Parameters:
// - lineNo:    Pointer to the current line number
// - pos:       Pointer to the current position
// - firstLine: Whether this is the first line of the comment
// - multiline: Whether this comment has already been identified as multi-line
//
// Returns:
// - ateClose: Whether the closing delimiter was consumed
// - err:      Any error that occurred during parsing
func (tc *TemplateClassifier) eatFullComment(lineNo *int, pos *int, firstLine bool, multiline bool) (ateClose bool, err error) {
	if tc.char(lineNo, pos) == tc.closeDelim[0] {
		// Potential closing delimiter
		*pos++
		et := tc.pastEnd(lineNo, pos)
		switch et {
		case EOL:
			// Reached end of line - this is a multi-line comment
			multiline = true
			tc.addCommentSegment(*lineNo-1, multiline, MultilineBegin, firstLine, MultilineMiddle)
		case EOT:
			// Reached end of template - unclosed comment
			err = errors.Join(ErrUnclosedCommentTag,
				ErrArg(TemplateLineErrArg, tc.line(lineNo)),
			)
			goto end
		case NotAtEnd:
			// Check if this is the second char of a closing delimiter
			if tc.char(lineNo, pos) == tc.closeDelim[1] {
				// Found complete closing delimiter
				tc.addCommentSegment(*lineNo, multiline, MultilineMiddle, firstLine, MultilineEnd)
				*pos++
				ateClose = true
				goto end
			}
		}
		// Not a closing delimiter - backtrack
		*pos--
	}

	// Process text content until we find a potential closing delimiter
	for {
		*pos++
		et := tc.pastEnd(lineNo, pos)
		switch et {
		case EOL:
			// Reached end of line - this is a multi-line comment
			multiline = true
			tc.addCommentSegment(*lineNo-1, multiline, MultilineBegin, firstLine, MultilineMiddle)
		case EOT:
			// Reached end of template - unclosed comment
			err = errors.Join(ErrUnclosedCommentTag,
				ErrArg(TemplateContentErrArg, tc.template),
			)
			goto end
		case NotAtEnd:
			// Continue until we find the first char of a potential closing delimiter
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

// parseSections processes section tags that begin with '#' (like {{#section}}).
// Delegates to parseEnclosingBeginTag with SectionTag type.
func (tc *TemplateClassifier) parseSections(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, SectionTag)
}

// parseInvertedSections processes inverted section tags that begin with '^'
// (like {{^section}}). Inverted sections render content when the value is falsy.
// Delegates to parseEnclosingBeginTag with InvertedSectionTag type.
func (tc *TemplateClassifier) parseInvertedSections(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, InvertedSectionTag)
}

// parseClosings processes closing tags that begin with '/' (like {{/section}}).
// It checks that the closing tag matches the most recent opening tag on the
// stack.
func (tc *TemplateClassifier) parseClosings(lineNo *int, pos *int) (err error) {
	var identifier string
	var tag Tag

	*pos++ // Eat closing char '/'

	// Consume the section name until the closing delimiter
	identifier, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}

	// Pop the most recent tag from the stack and verify it matches
	tag = tc.tagStack.Pop()
	if tag.Identifier != identifier {
		template := tc.Template(lineNo)
		err = errors.Join(ErrMismatchedDelimiters,
			ErrArg(OpeningIdentifierErrArg, tag.Identifier),
			ErrArg(ClosingIdentifier, identifier),
			ErrArg(TemplateLineErrArg, template),
		)
	}

	tc.AddSegment(lineNo, MakeEndTagSegment(tag.Type))

end:
	return err
}

// parseVars processes regular variable tags (like {{var}}) and dot tags ({{.}}).
// Variable tags render the value from the context, with HTML escaping applied.
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

	// Special case for dot tag (refers to current context)
	switch identifier {
	case ".":
		tc.AddSegment(lineNo, MakeTagSegment(DotTag))
	default:
		tc.AddSegment(lineNo, MakeTagSegment(VarTag))
	}

end:
	return err
}

// parsePartials processes partial tags that begin with '>' (like {{>partial}}).
// Partials allow including other templates within the current template.
func (tc *TemplateClassifier) parsePartials(lineNo *int, pos *int) (err error) {
	*pos++ // Eat partial identifying character '>'

	// Eat partial identifier
	_, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	// Eat tag closing delimiters
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}
	tc.AddSegment(lineNo, MakeTagSegment(PartialTag))
end:
	return err
}

// parseParents processes parent tags that begin with '<' (like {{<parent}}).
// Parent tags are used for template inheritance patterns. Delegates to
// parseEnclosingBeginTag with ParentTag type.
func (tc *TemplateClassifier) parseParents(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, ParentTag)
}

// parseBlocks processes block tags that begin with '$' (like {{$block}}). Block
// tags work with parent tags to define named sections in templates. Delegates to
// parseEnclosingBeginTag with BlockTag type.
func (tc *TemplateClassifier) parseBlocks(lineNo *int, pos *int) (err error) {
	return tc.parseEnclosingBeginTag(lineNo, pos, BlockTag)
}

// parseSetDelimiters processes set delimiter tags (like {{=< >=}}). These tags
// use bracketing '=' characters and allow changing the default delimiters of
// '{{' and '}}' to custom ones instead.
func (tc *TemplateClassifier) parseSetDelimiters(lineNo *int, pos *int) (err error) {
	var delimBytes [16]byte // Max buffer size for delimiters
	var delimLen int
	var newOpen, newClose string

	// Parse the new delimiters
	for {
		err = tc.notPastEnd(lineNo, pos, func() (err error) {
			*pos++
			return err
		})
		if err != nil {
			goto end
		}
		c := tc.char(lineNo, pos)
		switch c {
		case ' ', '\t':
			// Handle whitespace between delimiters
			if delimLen == 0 {
				// Whitespace came before opening delimiter. Ignore it
				continue
			}
			if newOpen == "" {
				// Whitespace came after a delimiter was captured but before opening delimiter was saved. Save it.
				newOpen = string(delimBytes[:delimLen])
				delimLen = 0
				continue
			}
			if newClose == "" {
				// Whitespace came after a delimiter was captured but after opening delimiter was saved. Save closing delimiter instead.
				newClose = string(delimBytes[:delimLen])
				delimLen = 0
			}
			continue
		case '=':
			// Handle the closing '=' that marks the end of delimiter definition
			if newOpen == "" {
				err = errors.Join(ErrDelimitersCannotBeEmpty,
					ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
				)
				goto end
			}
			if newClose == "" {
				// We did not see whitespace before we saw a closing set Delimiter char so we
				// never previously set newClose
				newClose = string(delimBytes[:delimLen])
			}
			delimLen = 0
			*pos++ // Eat the '='

			// We've reached the end of our delimiter definition, process closing delimiter
			err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
			if err != nil {
				goto end
			}

			// Validate the new delimiters
			if newOpen == TripleBraceBegin {
				err = errors.Join(ErrTripleBraceCannotBeCustomDelimiter,
					ErrArg(DelimiterPosition, OpenDelimErrArg),
					ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
				)
				goto end
			}
			if newClose == TripleBraceEnd {
				err = errors.Join(ErrTripleBraceCannotBeCustomDelimiter,
					ErrArg(DelimiterPosition, CloseDelimErrArg),
					ErrArg(TemplateLineErrArg, tc.Template(lineNo)),
				)
				goto end
			}

			// Set the new delimiters
			tc.openDelim = newOpen
			tc.closeDelim = newClose
			tc.AddSegment(lineNo, MakeTagSegment(SetDelimiterTag))
			goto end
		default:
			// Accumulate delimiter characters
			delimBytes[delimLen] = c
			delimLen++
		}
	}
end:
	return err
}

// parseDots processes dot tags that use the '.' syntax (like {{.}}).
// Dot tags reference the current context directly.
func (tc *TemplateClassifier) parseDots(lineNo *int, pos *int) (err error) {
	*pos++ // Eat dot '.'

	// Eat tag closing delimiters
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}
	tc.AddSegment(lineNo, MakeTagSegment(DotTag))
end:
	return err
}

// parseEnclosingBeginTag processes the opening tag for all enclosing tags
// (sections, blocks, etc.). This is a shared implementation for tags that have
// both opening and closing parts. It handles the identifier and pushes the tag
// onto the tag stack for matching with the closing tag.
func (tc *TemplateClassifier) parseEnclosingBeginTag(lineNo *int, pos *int, tagType TagType) (err error) {
	var identifier string

	*pos++ // Eat enclosing tag identifying character, e.g. '#', '$', '^', etc.

	// Eat identifier for the tag
	identifier, err = tc.maybeEatIdentifier(lineNo, pos)
	if err != nil {
		goto end
	}

	// Eat tag closing delimiters
	err = tc.maybeEatTagClosingDelimiters(lineNo, pos, DelimiterOpen)
	if err != nil {
		goto end
	}

	// Add segment and push tag onto stack for matching with closing tag
	tc.AddSegment(lineNo, MakeBeginTagSegment(tagType))
	tc.tagStack.Push(Tag{
		Type:       tagType,
		Identifier: identifier,
	})
end:
	return err
}

// SetLineType sets the type of a line in the classifier.
// It ensures the line exists before setting its type.
func (tc *TemplateClassifier) SetLineType(lineNo *int, lt LineType) {
	tc.ensureLines(lineNo)
	tc.lines[*lineNo].Type = lt
}

// AddSegment adds a segment to a line in the classifier.
// It ensures the line exists before adding the segment.
func (tc *TemplateClassifier) AddSegment(lineNo *int, segment Segment) {
	tc.ensureLines(lineNo)
	tc.lines[*lineNo].Segments = append(tc.lines[*lineNo].Segments, segment)
}

// ensureLines ensures that the lines slice is large enough to accommodate the
// specified line number. If not, it resizes both the templateLines and lines
// slices.
func (tc *TemplateClassifier) ensureLines(lineNo *int) {
	nextLineNo := *lineNo + 1
	if nextLineNo > len(tc.lines) {
		tc.templateLines = make([]string, nextLineNo)
		tc.lines = make(Lines, nextLineNo)
	}
}

// line returns a pointer to the Line at the specified line number.
func (tc *TemplateClassifier) line(lineNo *int) *Line {
	return &tc.lines[*lineNo]
}

// Template returns the raw template string at the specified line number.
func (tc *TemplateClassifier) Template(lineNo *int) string {
	return tc.templateLines[*lineNo]
}

// LineType returns the LineType of the line at the specified line number.
func (tc *TemplateClassifier) LineType(lineNo *int) LineType {
	return tc.lines[*lineNo].Type
}

// TemplateSubstring returns a substring of the template line at the specified
// line number.
func (tc *TemplateClassifier) TemplateSubstring(lineNo *int, begin, end int) string {
	return tc.templateLines[*lineNo][begin:end]
}

// char returns the character at the specified position in the specified line.
func (tc *TemplateClassifier) char(lineNo, pos *int) byte {
	return tc.templateLines[*lineNo][*pos]
}

// lineLen returns the length of the template line at the specified line number.
func (tc *TemplateClassifier) lineLen(lineNo *int) int {
	return len(tc.templateLines[*lineNo])
}

// isEmptyLine checks if the line at the specified line number is empty.
func (tc *TemplateClassifier) isEmptyLine(lineNo *int) bool {
	return len(tc.templateLines[*lineNo]) == 0
}

// AtEOL checks if the position is exactly at the end of the line (not past it).
func (tc *TemplateClassifier) AtEOL(lineNo, pos *int) bool {
	return tc.lines[*lineNo].AtEOL(pos, tc.lineLen(lineNo))
}

// PastEOL checks if the position is past the end of the line.
func (tc *TemplateClassifier) PastEOL(lineNo, pos *int) bool {
	return tc.lines[*lineNo].PastEOL(pos, tc.lineLen(lineNo))
}

// LineSegments returns the segments for the line at the specified line number.
func (tc *TemplateClassifier) LineSegments(lineNo *int) (ss Segments) {
	return tc.lines[*lineNo].Segments
}

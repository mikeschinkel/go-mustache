package classifier

import (
	"fmt"
	"testing"
)

type TestCase struct {
	name     string
	template string
	want     Lines
	error    string
	comment  string
}

var (
	partialTagSegment           = MakeTagSegment(PartialTag)
	setDelimiterTagSegment      = MakeTagSegment(SetDelimiterTag)
	varTagSegment               = MakeTagSegment(VarTag)
	commentTagSegment           = MakeTagSegment(CommentTag)
	sectionBeginSegment         = MakeBeginTagSegment(SectionTag)
	blockBeginSegment           = MakeBeginTagSegment(BlockTag)
	sectionBeginInvertedSegment = MakeBeginTagSegment(InvertedSectionTag)
	parentBeginSegment          = MakeBeginTagSegment(ParentTag)
	sectionEndSegment           = MakeEndTagSegment(SectionTag)
	sectionEndInvertedSegment   = MakeEndTagSegment(InvertedSectionTag)
	parentEndSegment            = MakeEndTagSegment(ParentTag)
	blockEndSegment             = MakeEndTagSegment(BlockTag)
	whitespaceSegment           = MakeWhitespaceSegment()
	textContentSegment          = MakeTextContentSegment()
	tripleBraceUnescapedSegment = MakeTagSegment(TripleBraceUnescaped)
	ampersandUnescapedSegment   = MakeTagSegment(AmpersandUnescaped)
	multilineCommentBegin       = MakeCommentTagSegment(MultilineBegin)
	multilineCommentMiddle      = MakeCommentTagSegment(MultilineMiddle)
	multilineCommentEnd         = MakeCommentTagSegment(MultilineEnd)

	partialTagSegments           = Segments{partialTagSegment}
	setDelimiterTagSegments      = Segments{setDelimiterTagSegment}
	varTagSegments               = Segments{varTagSegment}
	commentTagSegments           = Segments{commentTagSegment}
	sectionBeginSegments         = Segments{sectionBeginSegment}
	sectionEndSegments           = Segments{sectionEndSegment}
	sectionBeginInvertedSegments = Segments{sectionBeginInvertedSegment}
	sectionEndInvertedSegments   = Segments{sectionEndInvertedSegment}
	parentBeginSegments          = Segments{parentBeginSegment}
	parentEndSegments            = Segments{parentEndSegment}
	blockBeginSegments           = Segments{blockBeginSegment}
	blockEndSegments             = Segments{blockEndSegment}
	whitespaceSegments           = Segments{whitespaceSegment}
	textContentSegments          = Segments{textContentSegment}
	tripleBraceUnescapedSegments = Segments{tripleBraceUnescapedSegment}
	ampersandUnescapedSegments   = Segments{ampersandUnescapedSegment}
	multilineCommentBegins       = Segments{multilineCommentBegin}
	multilineCommentMiddles      = Segments{multilineCommentMiddle}
	multilineCommentEnds         = Segments{multilineCommentEnd}
)

func testsRunner(t *testing.T, tests []TestCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classifier := NewTemplateClassifier(tt.template)
			got, err := classifier.Classify()
			switch wantError(tt.error) {
			case NO:
				switch {
				case err != nil:
					t.Error(err.Error())
				case !got.Equal(tt.want):
					showMismatch(t, got.Normalize(), tt.want.Normalize(), tt.template, tt.comment)
				}
			case YES:
				switch {
				case err == nil:
					checkError(t, err, tt.error, tt.template, tt.comment)
				case err.Error() != tt.error:
					showMismatch(t, asError(err.Error()), asError(tt.error), tt.template, tt.comment)
				}
			}
		})
	}
}

const (
	YES = 'y'
	NO  = 'n'
)

func wantError(error string) byte {
	if error == "" {
		return NO
	}
	return YES
}

func checkError(t *testing.T, got error, want string, template, comment string) {
	if got == nil {
		t.Errorf("ERROR:"+
			"\n\tWant Error: %v"+
			"\n\tGot:        <No Error>"+
			"\n\tTemplate:   %q"+
			"\n\tComment:    %s\n",
			want,
			template,
			comment,
		)
	}
}

func showMismatch(t *testing.T, got, want any, template, comment string) {
	t.Errorf("ERROR:"+
		"\n\tGot:      %s"+
		"\n\tWant:     %s"+
		"\n\tTemplate: %q"+
		"\n\tComment:  %s\n",
		stringer(got),
		stringer(want),
		template,
		comment,
	)
}
func stringer(s any) string {
	switch ts := s.(type) {
	case string:
		return ts
	case fmt.Stringer:
		return ts.String()
	}
	return fmt.Sprintf("%v", s)
}
func asError(s string) string {
	return fmt.Sprintf("ERROR: `%s`", s)
}

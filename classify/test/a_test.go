package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/alexkappa/mustache/classify"
)

type (
	Segments = classify.Segments
	Lines    = classify.Lines
)

const (
	StandaloneLine = classify.StandaloneLine
	InlineLine     = classify.InlineLine
	TextLine       = classify.TextLine
	WhitespaceLine = classify.WhitespaceLine
	EmptyLine      = classify.EmptyLine
)

var (
	NewTemplateClassifier  = classify.NewTemplateClassifier
	MakeBeginTagSegment    = classify.MakeBeginTagSegment
	MakeSegment            = classify.MakeSegment
	MakeTagSegment         = classify.MakeTagSegment
	MakeEndTagSegment      = classify.MakeEndTagSegment
	MakeWhitespaceSegment  = classify.MakeWhitespaceSegment
	MakeTextContentSegment = classify.MakeTextContentSegment
	MakeCommentTagSegment  = classify.MakeCommentTagSegment
)

const (
	DotTag               = classify.DotTag
	PartialTag           = classify.PartialTag
	SetDelimiterTag      = classify.SetDelimiterTag
	VarTag               = classify.VarTag
	CommentTag           = classify.CommentTag
	SectionTag           = classify.SectionTag
	BlockTag             = classify.BlockTag
	InvertedSectionTag   = classify.InvertedSectionTag
	ParentTag            = classify.ParentTag
	TripleBraceUnescaped = classify.TripleBraceUnescaped
	AmpersandUnescaped   = classify.AmpersandUnescaped
	CompleteTag          = classify.CompleteTag
	MultilineBegin       = classify.MultilineBegin
	MultilineMiddle      = classify.MultilineMiddle
	MultilineEnd         = classify.MultilineEnd
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
			got, err := NewTemplateClassifier(tt.template).Classify()
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
				case !errEqual(err, tt.error):
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

func stringer(s any) (out string) {
	switch ts := s.(type) {
	case string:
		out = ts
	case fmt.Stringer:

		out = ts.String()
	default:
		out = fmt.Sprintf("%v", s)
	}
	return stringNormalizer(out)
}
func asError(s string) string {
	return fmt.Sprintf("ERROR: `%s`", s)
}
func stringNormalizer(s string) string {
	return strings.Replace(s, "\n", "; ", -1)
}

func errEqual(got error, want string) bool {
	return stringNormalizer(got.Error()) == stringNormalizer(want)
}

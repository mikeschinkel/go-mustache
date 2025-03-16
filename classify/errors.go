package classify

import (
	"errors"
	"fmt"
)

// Error argument descriptors
const (
	TagOpenTypeErrArg       = "tag_open_type"
	TemplateLineErrArg      = "template_line"
	TagStackErrArg          = "tag_stack"
	OpenDelimErrArg         = "opening_delimiter"
	CloseDelimErrArg        = "closing_delimiter"
	DelimiterPosition       = "delimiter_position"
	SegmentErrArg           = "segment"
	TemplateContentErrArg   = "template_content"
	OpeningIdentifierErrArg = "opening_identifier"
	ClosingIdentifier       = "closing_identifier"
	DelimiterTypeErrArg     = "delimiter_type"
	TripleBraceErrArg       = "triple_brace_delimiter"
	StandardDelimiterErrArg = "standard_delimiter"
)

// Sentinel errors
var (
	ErrInvalidTagOpeningType              = errors.New("invalid tag opening type")
	ErrUnclosedTags                       = errors.New("unclosed tags")
	ErrMismatchedDelimiters               = errors.New("mismatched opening and closing delimiters")
	ErrInvalidTagFormat                   = errors.New("invalid tag format")
	ErrMissingTagIdentifier               = errors.New("tag identifier is missing")
	ErrInvalidClosingTag                  = errors.New("invalid closing tag")
	ErrUnexpectedEOL                      = errors.New("unexpected end of line (EOL)")
	ErrUnexpectedEOT                      = errors.New("unexpected end of template (EOT)")
	ErrUnclosedCommentTag                 = errors.New("unclosed comment tag")
	ErrDelimitersCannotBeEmpty            = errors.New("delimiters cannot be empty")
	ErrTripleBraceCannotBeCustomDelimiter = errors.New("triple braces cannot be used as custom delimiters")
	ErrSegmentTypeMismatch                = errors.New("unexpected mismatch between segment type and tag type")
)

// ErrArg creates an error with a named argument for context
func ErrArg(name string, value any) error {
	return fmt.Errorf("%s=%v", name, value)
}

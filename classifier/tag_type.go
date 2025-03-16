package classifier

type TagTypes []TagType

// TagType represent the type of tag for a segment
type TagType uint8 // Now fits into 4 bits

func (tt TagType) IsValid() (valid bool) {
	switch tt {
	case NotApplicable, NonEnclosingTagType, IgnoredTagType:
		// accept zero value for valid
	default:
		valid = true
	}
	return valid
}

const (
	NotApplicable        TagType = iota
	AmpersandUnescaped           // {{&tag}}:       Unescaped tag with ampersand syntax: {{&tag}}
	BlockTag                     // {{$block}}:     Block definition tag with dollar syntax: {{$block}}
	CommentTag                   // {{! comment }}: Comment tag with exclamation point syntax: {{! Comment goes here}}
	SetDelimiterTag              // {{=< >=}}:      Delimiter setting tag with equals syntax: {{=< >=}}
	DotTag                       // {{.}}:          Current context tag with dot syntax: {{.}}
	InvertedSectionTag           // {{^section}}:   Inverted section tag with caret syntax: {{^section}}
	ParentTag                    // {{/parent}}:    Parent template tag end for inheritance with slash syntax: {{/parent}}
	PartialTag                   // {{>partial}}:   Partial inclusion tag with greater-than syntax: {{>partial}}
	VarTag                       // {{var}}:        Variable tag {{var}}
	SectionTag                   // {{#section}}:   Section tag with hash syntax: {{#section}}
	TripleBraceUnescaped         // {{{tag}}}:      Unescaped tag with triple brace syntax: {{{tag}}}
	NonEnclosingTagType          // Used when a tag type is needed but is not a enclosing tag type
	IgnoredTagType               // Used by Normalize() so that String() will return ""
)

var enclosingTagTypes = map[TagType]struct{}{
	BlockTag:           {},
	InvertedSectionTag: {},
	ParentTag:          {},
	SectionTag:         {},
}

//goland:noinspection GoUnusedFunction
func isEnclosingTagType(tt TagType) bool {
	_, ok := enclosingTagTypes[tt]
	return ok
}

// String returns a human-readable representation of Tag for debugging
func (tt TagType) String() string {
	switch tt {
	case AmpersandUnescaped:
		return "AmpersandUnescaped"
	case BlockTag:
		return "BlockTag"
	case CommentTag:
		return "CommentTag"
	case SetDelimiterTag:
		return "SetDelimiterTag"
	case DotTag:
		return "DotTag"
	case InvertedSectionTag:
		return "InvertedSectionTag"
	case ParentTag:
		return "ParentTag"
	case PartialTag:
		return "PartialTag"
	case VarTag:
		return "VarTag"
	case SectionTag:
		return "SectionTag"
	case TripleBraceUnescaped:
		return "TripleBraceUnescaped"
	case NonEnclosingTagType:
		return "NonEnclosingTagType"
	case NotApplicable, IgnoredTagType:
		return ""
	default:
		return "UnknownTagType"
	}
}

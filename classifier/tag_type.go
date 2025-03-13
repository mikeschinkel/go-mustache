package classifier

type TagTypes []TagType
type TagType int

func (tt TagType) IsValid() (valid bool) {
	switch tt {
	case BlockTag, InvertedSectionTag, ParentTag, SectionTag:
		// Enclosing tags
		fallthrough
	case AmpersandUnescaped, DelimiterTag, DotTag, PartialTag, VarTag, CommentTag, TripleBraceUnescaped:
		// Non-enclosing tags
		valid = true
	default:
	}
	return valid
}

const (
	NotApplicable        TagType = iota
	AmpersandUnescaped           // Unescaped tag with ampersand syntax: {{&tag}}
	BlockTag                     // Block definition tag with dollar syntax: {{$block}}
	CommentTag                   // Comment tag with exclamation point syntax: {{! Comment goes here}}
	DelimiterTag                 // Delimiter setting tag with equals syntax: {{=< >=}}
	DotTag                       // Current context tag with dot syntax: {{.}}
	InvertedSectionTag           // Inverted section tag with caret syntax: {{^section}}
	ParentTag                    // Parent template tag end for inheritance with slash syntax: {{/parent}}
	PartialTag                   // Partial inclusion tag with greater-than syntax: {{>partial}}
	VarTag                       // Variable tag {{var}}
	SectionTag                   // Section tag with hash syntax: {{#section}}
	TripleBraceUnescaped         // Unescaped tag with triple brace syntax: {{{tag}}}
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
	case DelimiterTag:
		return "DelimiterTag"
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
	case NotApplicable:
		return ""
	case IgnoredTagType:
		return ""
	default:
		return "UnknownTagType"
	}
}

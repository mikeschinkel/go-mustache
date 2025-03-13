package classifier

type OpenTagType int8

const (
	InvalidOpenTag OpenTagType = iota
	TripleBraceOpen
	DelimiterOpen
)

func (ott OpenTagType) String() string {
	switch ott {
	case TripleBraceOpen:
		return "TripleBraceOpen"
	case DelimiterOpen:
		return "DelimiterOpen"
	case InvalidOpenTag:
		fallthrough
	default:
		return "InvalidOpenTag"
	}
}

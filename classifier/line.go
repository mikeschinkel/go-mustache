package classifier

import (
	"reflect"
)

type Lines []Line

func (ll Lines) Normalize() Lines {
	for i := range ll {
		ll[i].Template = ""
		if len(ll[i].Segments) == 0 && ll[i].Segments != nil {
			ll[i].Segments = nil
		}
	}
	return ll
}

func (ll Lines) Equal(lines Lines) bool {
	if len(ll) != len(lines) {
		return false
	}
	return reflect.DeepEqual(ll.Normalize(), lines.Normalize())
}

type Line struct {
	Template string
	Type     LineType
	Segments Segments
}

func (l *Line) PastEOL(pos *int) bool {
	return *pos >= len(l.Template)
}
func (l *Line) AtEOL(pos *int) bool {
	return *pos == len(l.Template)
}

func (l *Line) Len() int {
	return len(l.Template)
}

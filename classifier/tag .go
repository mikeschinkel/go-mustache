package classifier

import (
	"fmt"
)

type Tag struct {
	Identifier string
	Type       TagType
}

func (t Tag) String() string {
	return fmt.Sprintf("[tag: {name: '%s', type: '%s'}]", t.Identifier, t.Type)
}

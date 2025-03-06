package mustache

import (
	"reflect"
	"slices"
)

type LineNoTypes []LineTypes

func (ts LineNoTypes) Equal(types LineNoTypes) bool {
	lntss := [2]LineNoTypes{ts, types}
	for i, lnts := range lntss {
		for j, lineTypes := range lnts {
			if len(lineTypes) <= 1 {
				continue
			}
			slices.Sort(lnts[j])
			// Assigning back to parent here happens far less often than outside this
			// loop, and without the sort there is no need to assign.
			lntss[i] = lnts
		}
	}
	return reflect.DeepEqual(lntss[0], lntss[1])
}

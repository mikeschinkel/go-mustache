package mustache

import (
	"reflect"
	"slices"
)

type LineTypes []LineAspects

func (ts LineTypes) Equal(types LineTypes) bool {
	lntss := [2]LineTypes{ts, types}
	for i, lnts := range lntss {
		for j, LineAspects := range lnts {
			if len(LineAspects) <= 1 {
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

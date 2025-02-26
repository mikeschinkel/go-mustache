package mustache

import (
	"errors"
)

type MultiErr []error

func (errs *MultiErr) Err() error {
	return errors.Join(*errs...)
}
func (errs *MultiErr) Add(err error) {
	if err == nil {
		return
	}
	*errs = append(*errs, err)
}

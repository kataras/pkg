package structcopy

import (
	"github.com/jinzhu/copier"
)

// Copy copies the fields of "src" to the "dst"
// It checks the type and the name, it supports embedded struct's fields and interfaces to struct matching as well.
//
// It just calls the jinzhu/copier#Copy method.
func Copy(dst interface{}, src interface{}) error {
	return copier.Copy(dst, src)
}

package starlark_warning

import (
	"fmt"
)

const WarningConstant = "[WARN]:"

type DeprecationDate struct {
	Day   int
	Month int
	Year  int
}

func (deprecationDate DeprecationDate) GetFormattedDate() string {
	return fmt.Sprintf("%v/%v/%v", deprecationDate.Day, deprecationDate.Month, deprecationDate.Year)
}

type DeprecationNotice struct {
	deprecationDate DeprecationDate
	mitigation      string
}

// Deprecation DeprecationDate - date when the field or the instruction will be deprecated
// mitigation - what is the alternative way to do it
// for example: please use `xyz` instead - for more info check out the docs <link>
// no:lint
func Deprecation(deprecationDate DeprecationDate, mitigation string) *DeprecationNotice {
	return &DeprecationNotice{
		deprecationDate: deprecationDate,
		mitigation:      mitigation,
	}
}

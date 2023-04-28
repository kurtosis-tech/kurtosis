package starlark_warning

import (
	"fmt"
)

const WarningConstant = "[WARN]:"

// DeprecationDate TODO: check if the date is valid
type DeprecationDate struct {
	Day   int
	Month int
	Year  int
}

func (deprecationDate *DeprecationDate) GetFormattedDate() string {
	return fmt.Sprintf("%v/%v/%v", deprecationDate.Day, deprecationDate.Month, deprecationDate.Year)
}

// DeprecationNotice
//TODO: enforce that these fields are required
// give examples for good mitigation examples; currently it's free form for folks to start using right away
type DeprecationNotice struct {
	deprecationDate DeprecationDate
	mitigation      string
}

func (deprecationNotice *DeprecationNotice) GetDeprecatedDate() string {
	return deprecationNotice.deprecationDate.GetFormattedDate()
}

func (deprecationNotice *DeprecationNotice) GetMitigation() string {
	return deprecationNotice.mitigation
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

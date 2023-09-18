package starlark_warning

import (
	"fmt"
	"go.starlark.net/starlark"
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
// TODO: enforce that these fields are required
//
//	give examples for good mitigation examples; currently it's free form for folks to start using right away
type DeprecationNotice struct {
	deprecationDate                                    DeprecationDate
	mitigation                                         string
	shouldShowDeprecationNoticeBaseOnArgumentValueFunc func(value starlark.Value) bool
}

func (deprecationNotice *DeprecationNotice) GetDeprecatedDate() string {
	return deprecationNotice.deprecationDate.GetFormattedDate()
}

func (deprecationNotice *DeprecationNotice) GetMitigation() string {
	return deprecationNotice.mitigation
}

func (deprecationNotice *DeprecationNotice) GetMaybeShouldShowDeprecationNoticeBaseOnArgumentValueFunc() func(value starlark.Value) bool {
	return deprecationNotice.shouldShowDeprecationNoticeBaseOnArgumentValueFunc
}

func (deprecationNotice *DeprecationNotice) IsDeprecatedDateScheduled() bool {

	if deprecationNotice.deprecationDate.Day < 1 ||
		deprecationNotice.deprecationDate.Day > 31 ||
		deprecationNotice.deprecationDate.Month < 1 ||
		deprecationNotice.deprecationDate.Month > 12 ||
		deprecationNotice.deprecationDate.Year < 2023 {
		return false
	}

	return true
}

// Deprecation DeprecationDate - date when the field or the instruction will be deprecated
// mitigation - what is the alternative way to do it
// for example: please use `xyz` instead - for more info check out the docs <link>
// no:lint
func Deprecation(
	deprecationDate DeprecationDate,
	mitigation string,
	shouldShowDeprecationNoticeBaseOnArgumentValueFunc func(value starlark.Value) bool,
) *DeprecationNotice {
	return &DeprecationNotice{
		deprecationDate: deprecationDate,
		mitigation:      mitigation,
		shouldShowDeprecationNoticeBaseOnArgumentValueFunc: shouldShowDeprecationNoticeBaseOnArgumentValueFunc,
	}
}

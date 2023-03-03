package kurtosis_type_constructor

import "go.starlark.net/starlark"

type KurtosisValueType interface {
	starlark.HasAttrs
}

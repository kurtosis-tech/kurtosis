package builtin_argument

import (
	"go.starlark.net/starlark"
)

type KurtosisValueType interface {
	starlark.HasAttrs

	Copy() (KurtosisValueType, error)
}

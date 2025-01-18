package kurtosis_starlark_framework

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/binding_constructors"
)

// KurtosisBuiltinPosition represents the position of a given kurtosis instruction within a Starlark script.
type KurtosisBuiltinPosition struct {
	filename string
	line     int32
	column   int32
}

func NewKurtosisBuiltinPosition(filename string, line int32, column int32) *KurtosisBuiltinPosition {
	return &KurtosisBuiltinPosition{
		line:     line,
		column:   column,
		filename: filename,
	}
}

func (position *KurtosisBuiltinPosition) String() string {
	return fmt.Sprintf("%s[%d:%d]", position.filename, position.line, position.column)
}

func (position *KurtosisBuiltinPosition) ToAPIType() *kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition {
	return binding_constructors.NewStarlarkInstructionPosition(position.filename, position.line, position.column)
}

func (position *KurtosisBuiltinPosition) GetFilename() string {
	return position.filename
}

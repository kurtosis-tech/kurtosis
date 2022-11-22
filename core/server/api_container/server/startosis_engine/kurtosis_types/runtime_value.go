package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
	"strings"
)

const (
	runtimeValueTypeName = "runtime_value"
	resultAttr           = "result"
	codeAttr             = "code"
)

type RuntimeValue struct {
	result starlark.String
	code   starlark.String
}

func NewRuntimeValue(result starlark.String, code starlark.String) *RuntimeValue {
	return &RuntimeValue{
		result: result,
		code:   code,
	}
}

// String the starlark.Value interface
func (rv *RuntimeValue) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(runtimeValueTypeName + "(")
	buffer.WriteString(resultAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", rv.result))
	buffer.WriteString(codeAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", rv.code))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (rv *RuntimeValue) Type() string {
	return serviceTypeName
}

// Freeze implements the starlark.Value interface
func (rv *RuntimeValue) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (rv *RuntimeValue) Truth() starlark.Bool {
	return rv.result != "" && rv.code != ""
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (rv *RuntimeValue) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", serviceTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (rv *RuntimeValue) Attr(name string) (starlark.Value, error) {
	switch name {
	case resultAttr:
		return rv.result, nil
	case codeAttr:
		return rv.code, nil
	default:
		return nil, fmt.Errorf("'%v' has no attribute '%v'", serviceTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (rv *RuntimeValue) AttrNames() []string {
	return []string{resultAttr, codeAttr}
}

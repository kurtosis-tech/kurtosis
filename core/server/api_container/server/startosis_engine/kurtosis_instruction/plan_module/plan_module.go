package plan_module

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	planModuleName = "plan"
)

func PlanModule(kurtosisPlanInstructions []*starlark.Builtin) *starlarkstruct.Module {
	moduleBuiltins := starlark.StringDict{}
	for _, oldInstruction := range kurtosisPlanInstructions {
		moduleBuiltins[oldInstruction.Name()] = oldInstruction
	}

	return &starlarkstruct.Module{
		Name:    planModuleName,
		Members: moduleBuiltins,
	}
}

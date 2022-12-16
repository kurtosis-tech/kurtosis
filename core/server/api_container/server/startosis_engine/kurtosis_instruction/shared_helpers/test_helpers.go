package shared_helpers

import "go.starlark.net/starlark"

func NewStarlarkThread(name string) *starlark.Thread {
	return &starlark.Thread{
		Name:       name,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
	}
}

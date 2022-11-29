package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type RuntimeValueStore struct {
	recipeResultMap map[string]map[string]starlark.Comparable
}

func NewRuntimeValueStore() *RuntimeValueStore {
	return &RuntimeValueStore{
		recipeResultMap: make(map[string]map[string]starlark.Comparable),
	}
}

func (re *RuntimeValueStore) CreateValue() string {
	uuid, _ := uuid_generator.GenerateUUIDString()
	re.recipeResultMap[uuid] = nil
	return uuid
}

func (re *RuntimeValueStore) SetValue(uuid string, value map[string]starlark.Comparable) {
	re.recipeResultMap[uuid] = value
}

func (re *RuntimeValueStore) GetValue(uuid string) (map[string]starlark.Comparable, error) {
	value, found := re.recipeResultMap[uuid]
	if !found {
		return nil, stacktrace.NewError("Runtime UUID '%v' was not found", uuid)
	}
	if value == nil {
		return nil, stacktrace.NewError("Runtime UUID '%v' was found, but not set", uuid)
	}
	return value, nil
}

package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/uuid_generator"
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

func (re *RuntimeValueStore) CreateValue() (string, error) {
	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while generating uuid for runtime value")
	}
	re.recipeResultMap[uuid] = nil
	return uuid, nil
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

package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type RuntimeValueStore struct {
	serviceAssociatedValues map[service.ServiceName]string
	recipeResultMap         map[string]map[string]starlark.Comparable
}

func NewRuntimeValueStore() *RuntimeValueStore {
	return &RuntimeValueStore{
		serviceAssociatedValues: map[service.ServiceName]string{},
		recipeResultMap:         make(map[string]map[string]starlark.Comparable),
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

func (re *RuntimeValueStore) GetOrCreateValueAssociatedWithService(serviceName service.ServiceName) (string, error) {
	if uuid, found := re.serviceAssociatedValues[serviceName]; found {
		delete(re.recipeResultMap, uuid) // deleting old values so that they do not interfere until that are set again
		return uuid, nil
	}
	uuid, err := re.CreateValue()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating a simple runtime value")
	}
	re.serviceAssociatedValues[serviceName] = uuid
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

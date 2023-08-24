package runtime_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type RuntimeValueStore struct {
	recipeResultMap                   map[string]map[string]starlark.Comparable
	serviceAssociatedValuesRepository *serviceAssociatedValuesRepository
}

func CreateRuntimeValueStore(enclaveDb *enclave_db.EnclaveDB) (*RuntimeValueStore, error) {
	associatedValuesRepository, err := getOrCreateNewServiceAssociatedValuesRepository(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting or creating the new service associated values repository")
	}

	runtimeValueStore := &RuntimeValueStore{
		recipeResultMap:                   make(map[string]map[string]starlark.Comparable),
		serviceAssociatedValuesRepository: associatedValuesRepository,
	}

	return runtimeValueStore, nil
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

	exist, err := re.serviceAssociatedValuesRepository.Exist(serviceName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking if there are associated values for service '%s' in the service associated values repository", serviceName)
	}
	if exist {
		uuid, getErr := re.serviceAssociatedValuesRepository.Get(serviceName)
		if getErr != nil {
			return "", stacktrace.Propagate(err, "An error occurred getting associated values for service '%s'", serviceName)
		}
		if uuid != "" {
			delete(re.recipeResultMap, uuid) // deleting old values so that they do not interfere until that are set again
			return uuid, nil
		}
	}
	uuid, err := re.CreateValue()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating a simple runtime value")
	}

	if err := re.serviceAssociatedValuesRepository.Save(serviceName, uuid); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred saving associated values '%s' for service '%s' in the service associated values repository", uuid, serviceName)
	}

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

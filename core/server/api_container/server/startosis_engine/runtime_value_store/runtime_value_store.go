package runtime_value_store

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type RuntimeValueStore struct {
	starlarkValueSerde                *kurtosis_types.StarlarkValueSerde
	recipeResultRepository            *recipeResultRepository
	serviceAssociatedValuesRepository *serviceAssociatedValuesRepository
}

func CreateRuntimeValueStore(starlarkValueSerde *kurtosis_types.StarlarkValueSerde, enclaveDb *enclave_db.EnclaveDB) (*RuntimeValueStore, error) {
	associatedValuesRepository, err := getOrCreateNewServiceAssociatedValuesRepository(enclaveDb)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting or creating the service associated values repository")
	}

	recipeResultRepositoryObj, err := getOrCreateNewRecipeResultRepository(enclaveDb, starlarkValueSerde)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting or creating the recipe result repository")
	}

	runtimeValueStore := &RuntimeValueStore{
		starlarkValueSerde:                starlarkValueSerde,
		recipeResultRepository:            recipeResultRepositoryObj,
		serviceAssociatedValuesRepository: associatedValuesRepository,
	}

	return runtimeValueStore, nil
}

func (re *RuntimeValueStore) CreateValue() (string, error) {
	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while generating uuid for runtime value")
	}

	if err = re.recipeResultRepository.SaveKey(uuid); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred saving key UUID '%s' on the recipe result repository", uuid)
	}

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

func (re *RuntimeValueStore) SetValue(uuid string, value map[string]starlark.Comparable) error {
	if err := re.recipeResultRepository.Save(uuid, value); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving value '%+v' using UUID key '%s' into the recipe result repository", value, uuid)
	}
	return nil
}

func (re *RuntimeValueStore) GetValue(uuid string) (map[string]starlark.Comparable, error) {
	value, err := re.recipeResultRepository.Get(uuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting recipe result value with UUID key '%s'", uuid)
	}
	if len(value) == 0 {
		return nil, stacktrace.NewError("Runtime UUID '%v' was found, but not set", uuid)
	}

	return value, nil
}

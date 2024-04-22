package interpretation_time_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
)

type InterpretationTimeValueStore struct {
	serviceConfigValues map[service.ServiceName]*service.ServiceConfig
	serviceValues       *serviceInterpretationValueRepository
	serde               *kurtosis_types.StarlarkValueSerde
}

func CreateInterpretationTimeValueStore(enclaveDb *enclave_db.EnclaveDB, serde *kurtosis_types.StarlarkValueSerde) (*InterpretationTimeValueStore, error) {
	serviceValuesRepository, err := getOrCreateNewServiceInterpretationTimeValueRepository(enclaveDb, serde)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating interpretation time value store")
	}
	return &InterpretationTimeValueStore{
		serviceConfigValues: map[service.ServiceName]*service.ServiceConfig{},
		serviceValues:       serviceValuesRepository,
		serde:               serde}, nil
}

func (itvs *InterpretationTimeValueStore) PutService(name service.ServiceName, service *kurtosis_types.Service) error {
	if err := itvs.serviceValues.PutService(name, service); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding value '%v' for service '%v' to db", service, name)
	}
	return nil
}

func (itvs *InterpretationTimeValueStore) GetService(name service.ServiceName) (*kurtosis_types.Service, error) {
	serviceStarlark, err := itvs.serviceValues.GetService(name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching interpretation time value for '%v' from db", name)
	}
	return serviceStarlark, nil
}

func (itvs *InterpretationTimeValueStore) PutServiceConfig(name service.ServiceName, serviceConfig *service.ServiceConfig) {
	itvs.serviceConfigValues[name] = serviceConfig
}

func (itvs *InterpretationTimeValueStore) GetServiceConfig(name service.ServiceName) (*service.ServiceConfig, error) {
	serviceConfig, ok := itvs.serviceConfigValues[name]
	if !ok {
		return nil, stacktrace.NewError("Did not find service config for '%v' in interpretation time value store.", name)
	}
	return serviceConfig, nil
}

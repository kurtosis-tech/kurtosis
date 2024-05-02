package interpretation_time_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
)

type InterpretationTimeValueStore struct {
	serviceConfigValues    map[service.ServiceName]*service.ServiceConfig
	setServiceConfigValues map[service.ServiceName]*service.ServiceConfig
	serviceValues          *serviceInterpretationValueRepository
	serde                  *kurtosis_types.StarlarkValueSerde
}

func CreateInterpretationTimeValueStore(enclaveDb *enclave_db.EnclaveDB, serde *kurtosis_types.StarlarkValueSerde) (*InterpretationTimeValueStore, error) {
	serviceValuesRepository, err := getOrCreateNewServiceInterpretationTimeValueRepository(enclaveDb, serde)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating interpretation time value store")
	}
	return &InterpretationTimeValueStore{
		serviceConfigValues:    map[service.ServiceName]*service.ServiceConfig{},
		setServiceConfigValues: map[service.ServiceName]*service.ServiceConfig{},
		serviceValues:          serviceValuesRepository,
		serde:                  serde}, nil
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

func (itvs *InterpretationTimeValueStore) GetServices() ([]*kurtosis_types.Service, error) {
	servicesStarlark, err := itvs.serviceValues.GetServices()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching interpretation time service objects from db")
	}
	return servicesStarlark, nil
}

func (itvs *InterpretationTimeValueStore) RemoveService(name service.ServiceName) error {
	err := itvs.serviceValues.RemoveService(name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing interpretation time service object for service '%v'", name)
	}
	return nil
}

func (itvs *InterpretationTimeValueStore) PutServiceConfig(name service.ServiceName, serviceConfig *service.ServiceConfig) {
	itvs.serviceConfigValues[name] = serviceConfig
}

func (itvs *InterpretationTimeValueStore) GetServiceConfig(name service.ServiceName) (*service.ServiceConfig, error) {
	serviceConfig, ok := itvs.serviceConfigValues[name]
	if !ok {
		return nil, stacktrace.NewError("Did not find new service config for '%v' in interpretation time value store.", name)
	}
	return serviceConfig, nil
}

func (itvs *InterpretationTimeValueStore) SetServiceConfig(name service.ServiceName, serviceConfig *service.ServiceConfig) {
	itvs.setServiceConfigValues[name] = serviceConfig
}

func (itvs *InterpretationTimeValueStore) ExistsNewServiceConfigForService(name service.ServiceName) bool {
	_, doesConfigFromSetServiceInstructionExists := itvs.setServiceConfigValues[name]
	return doesConfigFromSetServiceInstructionExists
}

func (itvs *InterpretationTimeValueStore) GetNewServiceConfig(name service.ServiceName) (*service.ServiceConfig, error) {
	newServiceConfig, ok := itvs.setServiceConfigValues[name]
	if !ok {
		return nil, stacktrace.NewError("Did not find new service config for '%v' in interpretation time value store.", name)
	}
	delete(itvs.setServiceConfigValues, name)
	return newServiceConfig, nil
}

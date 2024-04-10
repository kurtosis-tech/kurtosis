package interpretation_time_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
)

type InterpretationTimeValueStore struct {
	serviceValues *serviceInterpretationValueRepository
	serde         *kurtosis_types.StarlarkValueSerde
}

func CreateInterpretationTimeValueStore(enclaveDb *enclave_db.EnclaveDB, serde *kurtosis_types.StarlarkValueSerde) (*InterpretationTimeValueStore, error) {
	serviceValuesRepository, err := getOrCreateNewServiceInterpretationTimeValueRepository(enclaveDb, serde)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating interpretation time value store")
	}
	return &InterpretationTimeValueStore{serviceValues: serviceValuesRepository, serde: serde}, nil
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

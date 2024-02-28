package interpretation_time_value_store

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	serviceInterpretationValueBucketName = []byte("service-interpretation-value")
	emptyValue                           = []byte{}
)

type serviceInterpretationValueRepository struct {
	enclaveDb          *enclave_db.EnclaveDB
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde
}

func getOrCreateNewServiceInterpretationTimeValueRepository(
	enclaveDb *enclave_db.EnclaveDB,
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde,
) (*serviceInterpretationValueRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(serviceInterpretationValueBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the bucket for the service interpretation time value repository")
		}
		logrus.Debugf("Service value interpretation time store bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building service interpretation time value repository")
	}

	repository := &serviceInterpretationValueRepository{
		enclaveDb:          enclaveDb,
		starlarkValueSerde: starlarkValueSerde,
	}

	return repository, nil
}

func (repository *serviceInterpretationValueRepository) PutService(name service.ServiceName, service *kurtosis_types.Service) error {
	logrus.Debugf("Saving service interpretation value '%v' for service with name '%v' to", service, name)
	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceInterpretationValueBucketName)

		serviceNameKey := getKey(name)
		serializedValue := repository.starlarkValueSerde.Serialize(service)

		// save it to disk
		if err := bucket.Put(serviceNameKey, []byte(serializedValue)); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving service interpretation time value '%v' for service '%v'", serializedValue, serviceNameKey)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving service interpretation time value '%v' for service '%v'", service, name)
	}
	logrus.Debugf("Succesfully saved service '%v'", name)
	return nil
}

func (repository *serviceInterpretationValueRepository) GetService(name service.ServiceName) (*kurtosis_types.Service, error) {
	logrus.Debugf("Getting service interpretation time value for service '%v'", name)
	var value *kurtosis_types.Service

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceInterpretationValueBucketName)

		serviceNameKey := getKey(name)

		serviceSerializedValue := bucket.Get(serviceNameKey)

		// check for existence
		if serviceSerializedValue == nil {
			return stacktrace.NewError("Service '%v' doesn't exist in the repository", name)
		}

		isEmptyValue := len(serviceSerializedValue) == len(emptyValue)

		serviceSerializedValueStr := string(serviceSerializedValue)

		// if an empty value was found we return an error
		if isEmptyValue {
			return stacktrace.NewError("An empty value was found for service '%v'; this is unexpected", name)
		}

		deserializedValue, interpretationErr := repository.starlarkValueSerde.Deserialize(serviceSerializedValueStr)
		if interpretationErr != nil {
			return stacktrace.Propagate(interpretationErr, "an error occurred while deserializing object associated with service '%v' in repository", name)
		}

		var ok bool
		value, ok = deserializedValue.(*kurtosis_types.Service)
		if !ok {
			return stacktrace.NewError("An error occurred converting the deserialized value '%v' into required internal type", deserializedValue)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting service '%v' from db", name)
	}
	logrus.Debugf("Successfully got value for '%v'", name)
	return value, nil

}

func getKey(name service.ServiceName) []byte {
	return []byte(name)
}

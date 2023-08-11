package service_registration

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	serviceRegistrationBucketName = []byte("service-registration-repository")
)

type ServiceRegistrationRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func GetOrCreateNewServiceRegistrationRepository(enclaveDb *enclave_db.EnclaveDB) (*ServiceRegistrationRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(serviceRegistrationBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the service registration database bucket")
		}
		logrus.Debugf("Service registration bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the service registration repository")
	}

	serviceRegistrationRepository := &ServiceRegistrationRepository{
		enclaveDb: enclaveDb,
	}

	return serviceRegistrationRepository, nil
}

func (repository *ServiceRegistrationRepository) GetAll() (map[service.ServiceName]*service.ServiceRegistration, error) {
	allServiceRegistrationsByServiceName := map[service.ServiceName]*service.ServiceRegistration{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		if err := bucket.ForEach(func(serviceNameKey, serviceRegistrationBytes []byte) error {
			serviceNameStr := string(serviceNameKey)
			serviceName := service.ServiceName(serviceNameStr)

			serviceRegistration, err := getServiceRegistrationFromBytes(serviceRegistrationBytes, serviceName)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred getting service registrations for service '%s' from service registrations bytes", serviceName)
			}
			allServiceRegistrationsByServiceName[serviceName] = serviceRegistration

			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "An error occurred while iterating the service registration repository to get all registrations")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all service registrations from the service registration repository")
	}

	return allServiceRegistrationsByServiceName, nil
}

// Get returns a service registration struct previously saved, never returns a nil struct if it doesn't fail
// Exist() method can used for checking existence instead
func (repository *ServiceRegistrationRepository) Get(serviceName service.ServiceName) (*service.ServiceRegistration, error) {
	var (
		serviceRegistration *service.ServiceRegistration
		err                 error
	)

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		serviceRegistration, err = getServiceRegistrationFromBucket(bucket, serviceName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service registration for service '%s' from bucket with name '%s'", serviceName, serviceRegistrationBucketName)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting service registration from the service registration repository")
	}
	return serviceRegistration, nil
}

func (repository *ServiceRegistrationRepository) Exist(serviceName service.ServiceName) (bool, error) {

	exist := false

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		serviceNameKey := getServiceNameKey(serviceName)
		bytes := bucket.Get(serviceNameKey)
		if bytes != nil {
			exist = true
		}
		return nil
	}); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while checking if service registration for service '%s' exist in service registration repository", serviceName)
	}

	return exist, nil
}

func (repository *ServiceRegistrationRepository) GetAllServiceNames() (map[service.ServiceName]bool, error) {
	serviceNames := map[service.ServiceName]bool{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		if err := bucket.ForEach(func(serviceNameKey, _ []byte) error {
			serviceNameStr := string(serviceNameKey)
			serviceName := service.ServiceName(serviceNameStr)
			serviceNames[serviceName] = true
			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "An error occurred while iterating the service registration repository to get the service names list")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all service names from the service registration repository")
	}
	return serviceNames, nil
}

func (repository *ServiceRegistrationRepository) GetAllEnclaveServiceRegistrations(
	enclaveUuid enclave.EnclaveUUID,
) (map[service.ServiceUUID]*service.ServiceRegistration, error) {
	allEnclaveServiceRegistrations := map[service.ServiceUUID]*service.ServiceRegistration{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		if err := bucket.ForEach(func(serviceNameKey, serviceRegistrationBytes []byte) error {
			serviceNameStr := string(serviceNameKey)
			serviceName := service.ServiceName(serviceNameStr)
			serviceRegistration, err := getServiceRegistrationFromBytes(serviceRegistrationBytes, serviceName)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred getting service registration from bytes for service '%s'", serviceName)
			}

			if serviceRegistration.GetEnclaveID() == enclaveUuid {
				serviceUuid := serviceRegistration.GetUUID()
				allEnclaveServiceRegistrations[serviceUuid] = serviceRegistration
			}
			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "An error occurred while iterating the service registration repository to get the service names list")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all service names from the service registration repository")
	}

	return allEnclaveServiceRegistrations, nil
}

func (repository *ServiceRegistrationRepository) Save(
	serviceRegistration *service.ServiceRegistration,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		serviceName := serviceRegistration.GetName()
		if err := saveServiceRegistrationIntoTheBucket(bucket, serviceName, serviceRegistration); err != nil {
			return stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s' in the service registration bucket", serviceRegistration, serviceName)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving service registration '%v' into the service registration repository", serviceRegistration)
	}
	return nil
}

func (repository *ServiceRegistrationRepository) UpdateStatus(
	serviceName service.ServiceName,
	newServiceStatus service.ServiceStatus,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		// get the object form db
		serviceRegistration, err := getServiceRegistrationFromBucket(bucket, serviceName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service registration for service '%s' from bucket with name '%s'", serviceName, serviceRegistrationBucketName)
		}

		// update the object
		serviceRegistration.SetStatus(newServiceStatus)

		// save the updated object
		if err := saveServiceRegistrationIntoTheBucket(bucket, serviceName, serviceRegistration); err != nil {
			return stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s' in the service registration bucket", serviceRegistration, serviceName)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while updating the service status '%s' into the service registration repository for service '%s'", newServiceStatus, serviceName)
	}
	return nil
}

func (repository *ServiceRegistrationRepository) UpdateConfig(
	serviceName service.ServiceName,
	newServiceConfig *service.ServiceConfig,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		// get the object form db
		serviceRegistration, err := getServiceRegistrationFromBucket(bucket, serviceName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service registration for service '%s' from bucket with name '%s'", serviceName, serviceRegistrationBucketName)
		}

		// update the object
		serviceRegistration.SetConfig(newServiceConfig)

		// save the updated object
		if err := saveServiceRegistrationIntoTheBucket(bucket, serviceName, serviceRegistration); err != nil {
			return stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s' in the service registration bucket", serviceRegistration, serviceName)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while updating the service config '%+v' into the service registration repository for service '%s'", newServiceConfig, serviceName)
	}
	return nil
}

func (repository *ServiceRegistrationRepository) UpdateStatusAndConfig(
	serviceName service.ServiceName,
	newServiceStatus service.ServiceStatus,
	newServiceConfig *service.ServiceConfig,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		// get the object form db
		serviceRegistration, err := getServiceRegistrationFromBucket(bucket, serviceName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service registration for service '%s' from bucket with name '%s'", serviceName, serviceRegistrationBucketName)
		}

		// update the object
		serviceRegistration.SetStatus(newServiceStatus)
		serviceRegistration.SetConfig(newServiceConfig)

		// save the updated object
		if err := saveServiceRegistrationIntoTheBucket(bucket, serviceName, serviceRegistration); err != nil {
			return stacktrace.Propagate(err, "An error occurred saving service registration '%+v' for service '%s' in the service registration bucket", serviceRegistration, serviceName)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while updating the service status '%s' and service config '%+v' into the service registration repository for service '%s'", newServiceStatus, newServiceConfig, serviceName)
	}
	return nil
}

func (repository *ServiceRegistrationRepository) Delete(serviceName service.ServiceName) error {
	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceRegistrationBucketName)

		serviceNameKey := getServiceNameKey(serviceName)
		if err := bucket.Delete(serviceNameKey); err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting a service registration using with key '%v' from the service registration bucket", serviceNameKey)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while deleting the service registration for service '%s' from the service registration repository", serviceName)
	}
	return nil
}

func getServiceRegistrationFromBucket(bucket *bolt.Bucket, serviceName service.ServiceName) (*service.ServiceRegistration, error) {

	serviceNameKey := getServiceNameKey(serviceName)

	// first get the bytes
	serviceRegistrationBytes := bucket.Get(serviceNameKey)

	serviceRegistration, err := getServiceRegistrationFromBytes(serviceRegistrationBytes, serviceName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting service registrations for service '%s' from service registrations bytes", serviceName)
	}

	return serviceRegistration, nil
}

func getServiceRegistrationFromBytes(serviceRegistrationBytes []byte, serviceName service.ServiceName) (*service.ServiceRegistration, error) {
	// check for existence
	if serviceRegistrationBytes == nil {
		return nil, stacktrace.NewError("Service registration for service '%s' does not exist on the service registration repository", serviceName)
	}

	serviceRegistration := &service.ServiceRegistration{}

	if err := json.Unmarshal(serviceRegistrationBytes, serviceRegistration); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling service registration")
	}

	return serviceRegistration, nil
}

func saveServiceRegistrationIntoTheBucket(
	bucket *bolt.Bucket,
	serviceName service.ServiceName,
	serviceRegistration *service.ServiceRegistration,
) error {
	jsonBytes, err := json.Marshal(serviceRegistration)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling service registration '%+v'", serviceRegistration)
	}

	serviceNameKey := getServiceNameKey(serviceName)
	// save it to disk
	if err := bucket.Put(serviceNameKey, jsonBytes); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving service registration '%+v' into the enclave db bucket", serviceRegistration)
	}

	return nil
}

func getServiceNameKey(serviceName service.ServiceName) []byte {
	return []byte(serviceName)
}

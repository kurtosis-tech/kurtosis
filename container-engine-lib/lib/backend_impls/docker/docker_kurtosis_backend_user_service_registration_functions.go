package docker

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/user_service_registration"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

func (backend *DockerKurtosisBackend) CreateUserServiceRegistration(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceId user_service_registration.ServiceID,
) (*user_service_registration.UserServiceRegistration, error) {
	backend.serviceRegistrationMutex.Lock()
	defer backend.serviceRegistrationMutex.Unlock()

	registrationsForEnclave, found := backend.serviceRegistrations[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Service registration information isn't being tracked for enclave '%v', which indicates that no free IP address provider was " +
				"provided for this address at construction time; this likely means that user service registration is being called where it " +
				"shouldn't be (i.e. outside the API container)",
			enclaveId,
		)
	}
	if _, found := registrationsForEnclave[serviceId]; found {
		return nil, stacktrace.NewError(
			"Received request to register service with ID '%v' in enclave '%v', but a service with that ID is already registered",
			serviceId,
			enclaveId,
		)
	}

	freeIpAddrProvider, found := backend.enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, stacktrace.NewError(
			"Received a request to register service with ID '%v' in enclave '%v', but no free IP address provider was " +
				"defined for this enclave; this likely means that the user registration request is being called where it shouldn't " +
				"be (i.e. outside the API container)",
			serviceId,
			enclaveId,
		)
	}

	ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting a free IP address to give to service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	shouldFreeIp := true
	defer func() {
		if shouldFreeIp {
			freeIpAddrProvider.ReleaseIpAddr(ipAddr)
		}
	}()

	guid := user_service_registration.UserServiceRegistrationGUID(fmt.Sprintf(
		"%v--%v--%v",
		enclaveId,
		serviceId,
		time.Now().Unix(),
	))

	registration := user_service_registration.NewUserServiceRegistration(guid, enclaveId, serviceId, ipAddr)

	registrationsForEnclave[serviceId] = registration
	shouldDeleteRegistrationEntry := true
	func () {
		if shouldDeleteRegistrationEntry {
			delete(registrationsForEnclave, serviceId)
		}
	}()

	shouldFreeIp = false
	shouldDeleteRegistrationEntry = false
	return registration, nil
}

func (backend *DockerKurtosisBackend) GetUserServiceRegistrations(ctx context.Context, filters *user_service_registration.UserServiceRegistrationFilters) (map[user_service_registration.ServiceID]*user_service_registration.UserServiceRegistration, error) {
	backend.serviceRegistrationMutex.Lock()
	defer backend.serviceRegistrationMutex.Unlock()

	result := map[user_service_registration.ServiceID]*user_service_registration.UserServiceRegistration{}
	for enclaveId, serviceRegistrationsForEnclave := range backend.serviceRegistrations {
		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[enclaveId]; !found {
				continue
			}
		}

		for serviceId, serviceRegi


	}
}

func (backend *DockerKurtosisBackend) DestroyUserServiceRegistration(ctx context.Context, filters *user_service_registration.UserServiceRegistrationFilters) (resultSuccessfulServiceIds map[user_service_registration.ServiceID]bool, resultErroredServiceIds map[user_service_registration.ServiceID]error, resultErr error) {
	//TODO implement me
	panic("implement me")
}
package user_service_functions

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
	"time"
)

// Registers a user service for each given serviceID, allocating each an IP and ServiceGUID
func registerUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceIDs map[service.ServiceID]bool,
	serviceRegistrations map[enclave.EnclaveID]map[service.ServiceGUID]*service.ServiceRegistration,
	serviceRegistrationMutex *sync.Mutex,
	enclaveFreeIpProviders map[enclave.EnclaveID]*lib.FreeIpAddrTracker) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
	serviceRegistrationMutex.Lock()
	defer serviceRegistrationMutex.Unlock()
	successfulServicesPool := map[service.ServiceID]*service.ServiceRegistration{}
	failedServicesPool := map[service.ServiceID]error{}

	// If no service IDs passed in, return empty maps
	if len(serviceIDs) == 0 {
		return successfulServicesPool, failedServicesPool, nil
	}

	freeIpAddrProvider, found := enclaveFreeIpProviders[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"Received a request to register services in enclave '%v', but no free IP address provider was "+
				"defined for this enclave; this likely means that the registration request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveId,
		)
	}

	registrationsForEnclave, found := serviceRegistrations[enclaveId]
	if !found {
		return nil, nil, stacktrace.NewError(
			"No service registrations are being tracked for enclave '%v'; this likely means that the registration request is being called where it shouldn't "+
				"be (i.e. outside the API container)",
			enclaveId,
		)
	}

	successfulRegistrations := map[service.ServiceID]*service.ServiceRegistration{}
	failedRegistrations := map[service.ServiceID]error{}
	for serviceID, _ := range serviceIDs {
		ipAddr, err := freeIpAddrProvider.GetFreeIpAddr()
		if err != nil {
			failedRegistrations[serviceID] = stacktrace.Propagate(err, "An error occurred getting a free IP address to give to service '%v' in enclave '%v'", serviceID, enclaveId)
			continue
		}
		shouldFreeIp := true
		defer func() {
			if shouldFreeIp {
				freeIpAddrProvider.ReleaseIpAddr(ipAddr)
			}
		}()

		guid := service.ServiceGUID(fmt.Sprintf(
			"%v-%v",
			serviceID,
			time.Now().Unix(),
		))
		registration := service.NewServiceRegistration(
			serviceID,
			guid,
			enclaveId,
			ipAddr,
		)

		registrationsForEnclave[guid] = registration
		shouldRemoveRegistration := true
		defer func() {
			if shouldRemoveRegistration {
				delete(registrationsForEnclave, guid)

			}
		}()

		shouldFreeIp = false
		shouldRemoveRegistration = false
		successfulRegistrations[serviceID] = registration
	}

	// Add operations to their respective pools
	for serviceID, serviceRegistration := range successfulRegistrations {
		successfulServicesPool[serviceID] = serviceRegistration
	}

	for serviceID, serviceErr := range failedRegistrations {
		failedServicesPool[serviceID] = serviceErr
	}

	return successfulRegistrations, failedRegistrations, nil
}
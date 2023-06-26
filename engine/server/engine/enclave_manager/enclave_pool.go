package enclave_manager

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	oneEnclave = 1

	// This is the same default value used in the `kurtosis enclave add` CLI command
	defaultApiContainerLogLevel = logrus.DebugLevel

	// The partitioning feature is disabled by default because the Enclave Pool
	// is enabled only for K8s and the network partitioning feature is not implemented yet
	defaultIsPartitioningEnabled = false

	maxRetryAfterFails = 60

	timeBetweenRetries = time.Second * 1
)

type EnclavePool struct {
	kurtosisBackend  backend_interface.KurtosisBackend
	enclaveCreator   *EnclaveCreator
	idleEnclavesChan chan *kurtosis_engine_rpc_api_bindings.EnclaveInfo
	engineVersion    string
}

func CreateEnclavePool(
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveCreator *EnclaveCreator,
	poolSize uint8,
	engineVersion string,
) *EnclavePool {

	// The amount of idle enclaves is equal to the chan capacity + one enclave (this one will be waiting in the queue until the channel is unblocked)
	chanCapacity := poolSize - oneEnclave

	idleEnclavesChan := make(chan *kurtosis_engine_rpc_api_bindings.EnclaveInfo, chanCapacity)

	enclavePool := &EnclavePool{
		kurtosisBackend:  kurtosisBackend,
		enclaveCreator:   enclaveCreator,
		idleEnclavesChan: idleEnclavesChan,
		engineVersion:    engineVersion,
	}

	//TODO first iterate on all the existing enclaves in order to find idle enclaves already created

	go enclavePool.run()

	return enclavePool
}

func (pool *EnclavePool) GetEnclave(
	ctx context.Context,
	newEnclaveName string,
	engineVersion string,
	apiContainerVersion string,
	apiContainerLogLevel logrus.Level,
	isPartitioningEnabled bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	// TODO change the logLevel value ?? it's pending to check if it's possible
	// The enclaves in the pool are already configured with defaults params and is there no way to update
	// this config, so we have to check if the requested enclave params are equal to the enclaves stored
	// in the pool before returning one from here, if it not return nil
	if !areRequestedEnclaveParamsEqualToEnclaveInThePoolParams(
		engineVersion,
		apiContainerVersion,
		apiContainerLogLevel,
		isPartitioningEnabled,
	) {
		return nil, nil
	}

	enclaveInfo, ok := <-pool.idleEnclavesChan
	if !ok {
		return nil, stacktrace.NewError("A new enclave can't be returned from the pool because the internal channel is closed, it shouldn't happen; this is a bug in Kurtosis")
	}

	if err := pool.checkRunningAndRenameEnclave(ctx, enclave.EnclaveUUID(enclaveInfo.EnclaveUuid), newEnclaveName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while checking if the enclave with UUID '%v' is running and then renaming it to '%s'", enclaveInfo.EnclaveUuid, newEnclaveName)
	}

	logrus.Debugf("Returning enclave Info '%+v' for requested enclave name '%s'", enclaveInfo, newEnclaveName)

	return enclaveInfo, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
//TODO it's only removing the previous idle enclave, it's pending to implement the reusable feature
func (pool *EnclavePool) destroyPreviousIdleEnclaves(ctx context.Context) error {
	filters := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{},
		Statuses: map[enclave.EnclaveStatus]bool{
			enclave.EnclaveStatus_Running: true,
		},
	}

	enclaves, err := pool.kurtosisBackend.GetEnclaves(ctx, filters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves using filters '%+v'", filters)
	}

	// check if there are idle enclaves on the list
	idleEnclavesToRemove := map[enclave.EnclaveUUID]bool{}
	idleEnclavesToReuse := map[enclave.EnclaveUUID]bool{}

	for enclaveUUID, enclave := range enclaves {
		enclaveName := enclave.GetName()
		// is a idle enclave?
		if strings.HasPrefix(enclaveName, idleEnclaveNamePrefix) {
			// can be reused ?
			//TODO the reuse logic is not enable yet because we ned:
			//TODO 1- store the APIC version on the k8s and Docker labels in order to get it from the Kurtosis backend here
			//TODO 2- find a way to get the APIC private IP address because it will need it for filling the EnclaveInfo object
			//TODO 3- edit the CreationTime on the EnclaveInfo object before returning it from the pool
			if hasRequiredAPICVersion(enclave) && len(idleEnclavesToReuse) < pool.getPoolSize() {
				idleEnclavesToReuse[enclaveUUID] = true
				continue
			}

			// if it can't, should be deleted
			idleEnclavesToRemove[enclaveUUID] = true
		}
	}

	destroyEnclaveFilters := &enclave.EnclaveFilters{
		UUIDs:    idleEnclavesToRemove,
		Statuses: map[enclave.EnclaveStatus]bool{},
	}
	_, destroyEnclaveErrs, err := pool.kurtosisBackend.DestroyEnclaves(ctx, destroyEnclaveFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclaves using filters '%v'", destroyEnclaveFilters)
	}
	if len(destroyEnclaveErrs) > 0 {
		logrus.Errorf("Errors occurred removing the following enclaves")
		var removalErrorStrings []string
		for enclaveUUID, destroyEnclaveErr := range destroyEnclaveErrs {
			logrus.Errorf("Error '%v'", destroyEnclaveErr.Error())
			resultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR on Enclave %v <<<<<<<<<<<<<<<<<\n%v", enclaveUUID, destroyEnclaveErr.Error())
			removalErrorStrings = append(removalErrorStrings, resultErrStr)
		}
		joinedRemovalErrors := strings.Join(removalErrorStrings, errorDelimiter)
		return stacktrace.NewError("Following errors occurred while removing idle enclaves :\n%v", joinedRemovalErrors)
	}

	//TODO reuse this enclaves
	// TODO Check for channel capacity before adding it into the channel
	// pool.idleEnclavesChan <- newEnclaveInfo
	//logrus.Debugf("Enclave '%+v' was added intho the pool channel", newEnclaveInfo)

}

// Run will be executed in a sub-routine and will be in charge of:
// 1-
func (pool *EnclavePool) run() {

	defer close(pool.idleEnclavesChan)

	fails := 0

	for {

		newEnclaveInfo, err := pool.createNewEnclave()
		if err != nil {
			// Retry strategy if something fails
			logrus.Errorf("An error occurred creating a new idle enclave. Error\n%v", err)
			fails++
			if fails >= maxRetryAfterFails {
				//TODO log something
				break
			}
			time.Sleep(timeBetweenRetries)
			//TODO put some debug log
		}

		//TODO I'm nor really sure if we should store the Enclave Info here because the creation data
		//TODO will be old when we return the info or We can update the creation data before returnint it
		pool.idleEnclavesChan <- newEnclaveInfo
		logrus.Debugf("Enclave '%+v' was added intho the pool channel", newEnclaveInfo)
	}
}

func (pool *EnclavePool) createNewEnclave() (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	ctx := context.Background()

	enclaveName, err := GetRandomIdleEnclaveName()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred generating a random name for a new idle enclave.",
		)
	}

	apiContainerVersion := pool.engineVersion

	newEnclaveInfo, err := pool.enclaveCreator.CreateEnclave(
		ctx,
		apiContainerVersion,
		defaultApiContainerLogLevel,
		enclaveName,
		defaultIsPartitioningEnabled,
	)
	if err != nil {
		//TODO complete the message adding more values
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating idle enclave with name '%s'",
			enclaveName,
		)
	}

	logrus.Debugf("New idle enclave created '%+v'", newEnclaveInfo)
	return newEnclaveInfo, nil
}

// checkRunningAndRenameEnclave check if the enclave with UUID is still running.
// We do this because we store metadata in the Enclave Pool, and we have to
// be sure that the enclave referenced on this medata is still running, and
// we renamed it after that
func (pool *EnclavePool) checkRunningAndRenameEnclave(
	ctx context.Context,
	enclaveUUID enclave.EnclaveUUID,
	newEnclaveName string,
) error {

	filters := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUUID: true,
		},
		Statuses: map[enclave.EnclaveStatus]bool{
			enclave.EnclaveStatus_Running: true,
		},
	}

	enclaves, err := pool.kurtosisBackend.GetEnclaves(ctx, filters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves using filters '%+v'", filters)
	}
	enclavesLen := len(enclaves)

	if enclavesLen == 0 {
		return stacktrace.NewError("There is not any running enclave with UUID '%v', it could have been stopped or destroyed", enclaveUUID)
	}
	if enclavesLen > 1 {
		return stacktrace.NewError("Expected to find only one running enclave with UUID '%v', but '%v' were found, this is a bug in Kurtosis", enclaveUUID, enclavesLen)
	}

	if err := pool.kurtosisBackend.RenameEnclave(ctx, enclaveUUID, newEnclaveName); err != nil {
		return stacktrace.Propagate(err, "An error occurred renaming enclave with UUID '%v' to '%s'", enclaveUUID, newEnclaveName)
	}

	return nil
}

func (pool *EnclavePool) getPoolSize() int {
	poolSize := cap(pool.idleEnclavesChan) + oneEnclave
	return poolSize
}

func hasRequiredAPICVersion(enclave *enclave.Enclave) bool {

	//TODO this is not implemented yet because we have to do some pre work for storing
	//TODO the APIC version on the k8s and Docker labels in order to get it from the Kurtosis backend here
	return false
}

// The enclave pool feature is only available for Kubernetes so far, and it will be activated
// only if users require this when setting the pool-size value
func isEnclavePoolAllowedForThisConfig(
	poolSize uint8,
	kurtosisBackendType args.KurtosisBackendType,
) bool {

	if poolSize > 0 && kurtosisBackendType == args.KurtosisBackendType_Kubernetes {
		return true
	}

	return false
}

func areRequestedEnclaveParamsEqualToEnclaveInThePoolParams(
	engineVersion string,
	apiContainerVersion string,
	apiContainerLogLevel logrus.Level,
	isPartitioningEnabled bool,
) bool {

	if engineVersion == apiContainerVersion &&
		apiContainerLogLevel == defaultApiContainerLogLevel &&
		isPartitioningEnabled == defaultIsPartitioningEnabled {
		return true
	}

	return false
}

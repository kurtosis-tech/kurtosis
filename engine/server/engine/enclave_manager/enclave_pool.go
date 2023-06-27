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
	idleEnclavesChan chan enclave.EnclaveUUID
	engineVersion    string
}

// CreateEnclavePool will do the following:
// 1- Wil create a new enclave pool object or return an error
// 2- Will remove idle enclaves from previous engine runs
// 3- Will start a sub-routine in charge of filling the pool
func CreateEnclavePool(
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveCreator *EnclaveCreator,
	poolSize uint8,
	engineVersion string,
) (*EnclavePool, error) {

	// The amount of idle enclaves is equal to the chan capacity + one enclave (this one will be waiting in the queue until the channel is unblocked)
	chanCapacity := poolSize - oneEnclave

	idleEnclavesChan := make(chan enclave.EnclaveUUID, chanCapacity)

	enclavePool := &EnclavePool{
		kurtosisBackend:  kurtosisBackend,
		enclaveCreator:   enclaveCreator,
		idleEnclavesChan: idleEnclavesChan,
		engineVersion:    engineVersion,
	}

	//TODO the current implementation only removes the previous idle enclave, it's pending to implement the reusable feature
	//TODO the reuse logic is not enable yet because we ned to store the APIC version on the APIContainer object in container-engine-lib
	//TODO in order to using it for comparing it with the expected version

	// iterate on all the existing enclaves in order to find idle enclaves already created
	// and reuse or destroy them if these were created from old Kurtosis version
	if err := enclavePool.destroyPreviousIdleEnclaves(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying previous idle enclave before creating the enclave pool")
	}

	go enclavePool.run()

	return enclavePool, nil
}

func (pool *EnclavePool) GetEnclave(
	ctx context.Context,
	newEnclaveName string,
	engineVersion string,
	apiContainerVersion string,
	apiContainerLogLevel logrus.Level,
	isPartitioningEnabled bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	logrus.Debugf(
		"Requesting enclave from pool using params: engine version '%s', api container version '%s', api container log level '%s' and is partitioning enabled '%v'...",
		engineVersion,
		apiContainerVersion,
		apiContainerLogLevel,
		isPartitioningEnabled,
	)

	// TODO change the logLevel value ?? it's pending to check if it's possible
	// The enclaves in the pool are already configured with defaults params and there is no way to update
	// this config, so we have to check if the requested enclave params are equal to the enclaves stored
	// in the pool before returning one
	if !areRequestedEnclaveParamsEqualToEnclaveInThePoolParams(
		engineVersion,
		apiContainerVersion,
		apiContainerLogLevel,
		isPartitioningEnabled,
	) {
		logrus.Debugf("The requested enclave params are different from the enclave in the pool params")
		return nil, nil
	}

	enclaveUUID, ok := <-pool.idleEnclavesChan
	if !ok {
		return nil, stacktrace.NewError("A new enclave can't be returned from the pool because the internal channel is closed, it shouldn't happen; this is a bug in Kurtosis")
	}

	enclaveObj, err := pool.getRunningEnclave(ctx, enclaveUUID)
	if err != nil {
		logrus.Debugf("The idle enclave with UUID '%v' is not longer running or have been destroyed", enclaveUUID)
		return nil, stacktrace.Propagate(err, "An error occurred getting a running enclave with UUID '%v'", enclaveUUID)
	}

	if err := pool.kurtosisBackend.RenameEnclave(ctx, enclaveUUID, newEnclaveName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred renaming enclave with UUID '%v' to '%s'", enclaveUUID, newEnclaveName)
	}

	enclaveInfo, err := getEnclaveInfoForEnclave(ctx, pool.kurtosisBackend, enclaveObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave info for enclave '%v'", enclaveObj)
	}
	enclaveInfo.Name = newEnclaveName

	logrus.Debugf("Returning enclave Info '%+v' for requested enclave name '%s'", enclaveInfo, newEnclaveName)

	return enclaveInfo, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func (pool *EnclavePool) destroyPreviousIdleEnclaves() error {
	ctx := context.Background()

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

	idleEnclavesToRemove := map[enclave.EnclaveUUID]bool{}

	for enclaveUUID, enclaveObj := range enclaves {
		enclaveName := enclaveObj.GetName()
		// is it an idle enclave from a previous run?
		if strings.HasPrefix(enclaveName, idleEnclaveNamePrefix) {
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
		logrus.Errorf("Errors occurred removing the following enclaves...")
		var removalErrorStrings []string
		for enclaveUUID, destroyEnclaveErr := range destroyEnclaveErrs {
			logrus.Errorf("Enclave wit UUID '%v' error:\n %v", enclaveUUID, destroyEnclaveErr.Error())
			resultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR on Enclave %v <<<<<<<<<<<<<<<<<\n%v", enclaveUUID, destroyEnclaveErr.Error())
			removalErrorStrings = append(removalErrorStrings, resultErrStr)
		}
		logrus.Errorf("...you should have to manually remove all these errored enclaves.")
		errorSeparator := "\n"
		joinedRemovalErrors := strings.Join(removalErrorStrings, errorSeparator)
		return stacktrace.NewError("Following errors occurred while removing idle enclaves :\n%v", joinedRemovalErrors)
	}

	return nil
}

// run method is in charge of filling the enclave pool channel with new idle enclaves
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
				logrus.Debugf("The enclave pool sub-routine has been canceled after reaching the max retries value '%v'", maxRetryAfterFails)
				break
			}
			logrus.Debugf("Next retry will be executed on '%v' seconds", timeBetweenRetries.Seconds())
			time.Sleep(timeBetweenRetries)
		}
		enclaveUUID := enclave.EnclaveUUID(newEnclaveInfo.EnclaveUuid)

		pool.idleEnclavesChan <- enclaveUUID
		logrus.Debugf("Enclave with UUID '%v' was added intho the pool channel", enclaveUUID)
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
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating idle enclave with name '%s', api container version '%v' and container log level '%s'",
			enclaveName,
			apiContainerVersion,
			defaultApiContainerLogLevel,
		)
	}

	logrus.Debugf("New idle enclave created '%+v'", newEnclaveInfo)
	return newEnclaveInfo, nil
}

func (pool *EnclavePool) getRunningEnclave(ctx context.Context, enclaveUUID enclave.EnclaveUUID) (*enclave.Enclave, error) {
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
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters '%+v'", filters)
	}
	enclavesLen := len(enclaves)

	if enclavesLen == 0 {
		return nil, stacktrace.NewError("There is not any running enclave with UUID '%v', it could have been stopped or destroyed", enclaveUUID)
	}
	if enclavesLen > 1 {
		return nil, stacktrace.NewError("Expected to find only one running enclave with UUID '%v', but '%v' were found, this is a bug in Kurtosis", enclaveUUID, enclavesLen)
	}

	enclaveObj, found := enclaves[enclaveUUID]
	if !found {
		return nil, stacktrace.NewError("Expected to find an enclave with UUID '%v' in enclave map '%+v', but none was found; this is a bug in Kurtosis", enclaveUUID, enclaves)
	}

	return enclaveObj, nil
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

	// if the api container version is empty string means that will be executed with the default version
	// which is the same that the current engine version
	if apiContainerVersion == "" {
		apiContainerVersion = engineVersion
	}

	if engineVersion == apiContainerVersion &&
		apiContainerLogLevel == defaultApiContainerLogLevel &&
		isPartitioningEnabled == defaultIsPartitioningEnabled {
		return true
	}

	return false
}

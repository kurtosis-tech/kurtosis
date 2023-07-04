package enclave_manager

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"strings"
	"sync"
)

const (
	apiContainerListenGrpcPortNumInsideNetwork = uint16(7443)

	getRandomEnclaveIdRetries = uint16(5)

	validNumberOfUuidMatches = 1

	errorDelimiter = ", "

	enclaveNameNotFound = "Name Not Found"
)

// No changes on enclave manager

// TODO Move this to the KurtosisBackend to calculate!!
// Completeness enforced via unit test
var isContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused:     false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running:    true,
	types.ContainerStatus_Removing:   false,
	types.ContainerStatus_Dead:       false,
	types.ContainerStatus_Created:    false,
	types.ContainerStatus_Exited:     false,
}

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// We use Docker as our backing datastore, but it has tons of race conditions so we use this mutex to ensure
	//  enclave modifications are atomic
	mutex *sync.Mutex

	kurtosisBackend                           backend_interface.KurtosisBackend
	apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier

	// this is a stop gap solution, this would be stored and retrieved from the DB in the future
	// we go with the GRPC type as it is just used by the engine server service
	// this is an append only list
	allExistingAndHistoricalIdentifiers []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers

	enclaveCreator *EnclaveCreator
	enclavePool    *EnclavePool
}

func CreateEnclaveManager(
	kurtosisBackend backend_interface.KurtosisBackend,
	kurtosisBackendType args.KurtosisBackendType,
	apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier,
	engineVersion string,
	poolSize uint8,
) (*EnclaveManager, error) {
	enclaveCreator := newEnclaveCreator(kurtosisBackend, apiContainerKurtosisBackendConfigSupplier)

	var (
		err         error
		enclavePool *EnclavePool
	)

	// The enclave pool feature is only available for Kubernetes so far
	if kurtosisBackendType == args.KurtosisBackendType_Kubernetes {
		enclavePool, err = CreateEnclavePool(kurtosisBackend, enclaveCreator, poolSize, engineVersion)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating enclave pool with pool-size '%v' and engine version '%v'", poolSize, engineVersion)
		}
	}

	enclaveManager := &EnclaveManager{
		mutex:           &sync.Mutex{},
		kurtosisBackend: kurtosisBackend,
		apiContainerKurtosisBackendConfigSupplier: apiContainerKurtosisBackendConfigSupplier,
		allExistingAndHistoricalIdentifiers:       []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{},
		enclaveCreator:                            enclaveCreator,
		enclavePool:                               enclavePool,
	}

	return enclaveManager, nil
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//
//	is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// If blank, will use the default
	engineVersion string,
	apiContainerImageVersionTag string,
	apiContainerLogLevel logrus.Level,
	//If blank, will use a random one
	enclaveName string,
	isPartitioningEnabled bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	var (
		enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo
		err         error
	)

	// TODO we could improve performance here by storing the names in memory for further requests
	allCurrentEnclaves, err := manager.kurtosisBackend.GetEnclaves(setupCtx, getAllEnclavesFilter())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for enclaves with name '%v'", enclaveName)
	}

	if enclaveName == autogenerateEnclaveNameKeyword {
		enclaveName = GetRandomEnclaveNameWithRetries(name_generator.GenerateNatureThemeNameForEnclave, allCurrentEnclaves, getRandomEnclaveIdRetries)
	}

	if isEnclaveNameInUse(enclaveName, allCurrentEnclaves) {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveName)
	}

	if err := validateEnclaveName(enclaveName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating enclave name '%v'", enclaveName)
	}

	if manager.enclavePool != nil {
		enclaveInfo, err = manager.enclavePool.GetEnclave(
			setupCtx,
			enclaveName,
			engineVersion,
			apiContainerImageVersionTag,
			apiContainerLogLevel,
			isPartitioningEnabled,
		)
		if err != nil {
			logrus.Errorf("An error occurred when trying to get an enclave from the enclave pool. Err:\n%v", err)
		}
	}

	if enclaveInfo == nil {
		enclaveInfo, err = manager.enclaveCreator.CreateEnclave(
			setupCtx,
			apiContainerImageVersionTag,
			apiContainerLogLevel,
			enclaveName,
			isPartitioningEnabled,
		)
		if err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating new enclave with name '%s' using api container image version '%s' and api container log level '%v'",
				enclaveName,
				apiContainerImageVersionTag,
				apiContainerLogLevel,
			)
		}
	}

	enclaveIdentifier := &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
		EnclaveUuid:   enclaveInfo.EnclaveUuid,
		Name:          enclaveInfo.Name,
		ShortenedUuid: enclaveInfo.ShortenedUuid,
	}
	manager.allExistingAndHistoricalIdentifiers = append(manager.allExistingAndHistoricalIdentifiers, enclaveIdentifier)

	return enclaveInfo, nil
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//
//	is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) GetEnclaves(
	ctx context.Context,
) (map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaves, err := manager.getEnclavesWithoutMutex(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclaves without the mutex")
	}

	// Transform map[enclave.EnclaveUUID]*EnclaveInfo -> map[string]*EnclaveInfo
	enclaveMapKeyedWithUuidStr := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for enclaveUuid, enclaveInfo := range enclaves {
		enclaveMapKeyedWithUuidStr[string(enclaveUuid)] = enclaveInfo
	}

	return enclaveMapKeyedWithUuidStr, nil
}

// StopEnclave
func (manager *EnclaveManager) StopEnclave(ctx context.Context, enclaveIdentifier string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveUuid, err := manager.getEnclaveUuidForIdentifierUnlocked(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching enclave uuid for identifier '%v'", enclaveIdentifier)
	}

	return manager.stopEnclaveWithoutMutex(ctx, enclaveUuid)
}

// DestroyEnclave
// TODO remove these notes - this should be working on active enclaves as well
// Destroys an enclave, deleting all objects associated with it in the container engine (containers, volumes, networks, etc.)
func (manager *EnclaveManager) DestroyEnclave(ctx context.Context, enclaveIdentifier string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveUuid, err := manager.getEnclaveUuidForIdentifierUnlocked(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while fetching enclave uuid for identifier '%v'", enclaveIdentifier)
	}

	enclaveDestroyFilter := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUuid: true,
		},
		Statuses: nil,
	}
	successfullyDestroyedEnclaves, erroredEnclaves, err := manager.kurtosisBackend.DestroyEnclaves(ctx, enclaveDestroyFilter)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the enclave")
	}
	if _, found := successfullyDestroyedEnclaves[enclaveUuid]; found {
		return nil
	}
	destructionErr, found := erroredEnclaves[enclaveUuid]
	if !found {
		return stacktrace.NewError("The requested enclave UUD '%v' for identifier '%v' wasn't found in the successfully-destroyed enclaves map, nor in the errors map; this is a bug in Kurtosis!", enclaveUuid, enclaveIdentifier)
	}
	return destructionErr
}

func (manager *EnclaveManager) Clean(ctx context.Context, shouldCleanAll bool) ([]*kurtosis_engine_rpc_api_bindings.EnclaveNameAndUuid, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	// TODO: Refactor with kurtosis backend
	var resultEnclaveNameAndUuids []*kurtosis_engine_rpc_api_bindings.EnclaveNameAndUuid

	// we prefetch the enclaves before deletion so that we have metadata
	enclavesForUuidNameMapping, err := manager.getEnclavesWithoutMutex(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Tried retrieving existing enclaves but failed")
	}

	successfullyRemovedEnclaveUuidStrs, removalErrors, err := manager.cleanEnclaves(ctx, shouldCleanAll)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves with shouldCleanAll set to '%v'", shouldCleanAll)
	}

	if len(removalErrors) > 0 {
		logrus.Errorf("Errors occurred removing the following enclaves")
		var removalErrorStrings []string
		for idx, err := range removalErrors {
			logrus.Errorf("Error '%v'", err.Error())
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, err.Error())
			removalErrorStrings = append(removalErrorStrings, indexedResultErrStr)
		}
		joinedRemovalErrors := strings.Join(removalErrorStrings, errorDelimiter)
		return nil, stacktrace.NewError("Following errors occurred while removing some enclaves :\n%v", joinedRemovalErrors)
	}

	if len(successfullyRemovedEnclaveUuidStrs) > 0 {
		logrus.Infof("Successfully removed the enclaves")
		sort.Strings(successfullyRemovedEnclaveUuidStrs)
		for _, successfullyRemovedEnclaveUuidStr := range successfullyRemovedEnclaveUuidStrs {
			nameAndUuid := &kurtosis_engine_rpc_api_bindings.EnclaveNameAndUuid{
				Uuid: successfullyRemovedEnclaveUuidStr,
				Name: enclaveNameNotFound,
			}
			// this should always be found; but we don't want to error if it isn't
			// we just use the default not found that we set above if we can't find the name
			enclave, found := enclavesForUuidNameMapping[enclave.EnclaveUUID(successfullyRemovedEnclaveUuidStr)]
			if found {
				nameAndUuid.Name = enclave.GetName()
			}
			resultEnclaveNameAndUuids = append(resultEnclaveNameAndUuids, nameAndUuid)
			logrus.Infof("Enclave Uuid '%v'", successfullyRemovedEnclaveUuidStr)
		}
	}

	return resultEnclaveNameAndUuids, nil
}

func (manager *EnclaveManager) GetEnclaveUuidForEnclaveIdentifier(ctx context.Context, enclaveIdentifier string) (enclave.EnclaveUUID, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	return manager.getEnclaveUuidForIdentifierUnlocked(ctx, enclaveIdentifier)
}

func (manager *EnclaveManager) GetExistingAndHistoricalEnclaveIdentifiers() ([]*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if len(manager.allExistingAndHistoricalIdentifiers) > 0 {
		return manager.allExistingAndHistoricalIdentifiers, nil
	}
	// either the engine got restarted or no enclaves have been created so far

	var enclaveIdentifiersResult []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers
	// TODO fix this - this is a hack while we persist enclave identifier information to disk
	// this is a hack that will only send enclaves that are still registered; removed or destroyed enclaves will not show up
	ctx := context.Background()
	allCurrentEnclavesToBackFillRestart, err := manager.getEnclavesWithoutMutex(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Found no registered enclaves in the in memory map; tried fetching them from backend but failed")
	}

	for _, enclave := range allCurrentEnclavesToBackFillRestart {
		identifiers := &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
			EnclaveUuid:   enclave.EnclaveUuid,
			Name:          enclave.Name,
			ShortenedUuid: enclave.ShortenedUuid,
		}
		enclaveIdentifiersResult = append(enclaveIdentifiersResult, identifiers)
	}

	return enclaveIdentifiersResult, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================

func getEnclaveApiContainerInformation(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclaveId enclave.EnclaveUUID,
) (
	kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo,
	error,
) {
	apiContainerByEnclaveIdFilter := getApiContainerByEnclaveIdFilter(enclaveId)
	enclaveApiContainers, err := kurtosisBackend.GetAPIContainers(ctx, apiContainerByEnclaveIdFilter)
	if err != nil {
		return 0, nil, nil, stacktrace.Propagate(err, "An error occurred getting the containers for enclave '%v'", enclaveId)
	}
	numOfFoundApiContainers := len(enclaveApiContainers)
	if numOfFoundApiContainers == 0 {
		return kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_NONEXISTENT,
			nil, nil, nil
	}
	if numOfFoundApiContainers > 1 {
		return 0, nil, nil, stacktrace.NewError("Expected to be able to find only one API container associated with enclave '%v', instead found '%v'",
			enclaveId, numOfFoundApiContainers)
	}

	apiContainer := getFirstApiContainerFromMap(enclaveApiContainers)

	resultApiContainerStatus, err := getApiContainerStatusFromContainerStatus(apiContainer.GetStatus())
	if err != nil {
		return 0, nil, nil, stacktrace.Propagate(err, "An error occurred getting the API container status for enclave '%v'", enclaveId)
	}
	resultApiContainerInfo := &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
		ContainerId:           "",
		IpInsideEnclave:       apiContainer.GetPrivateIPAddress().String(),
		GrpcPortInsideEnclave: uint32(apiContainer.GetPrivateGRPCPort().GetNumber()),
	}
	var resultApiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	if resultApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {

		var apiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
		if apiContainer.GetPublicIPAddress() != nil &&
			apiContainer.GetPublicGRPCPort() != nil {

			apiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
				IpOnHostMachine:       apiContainer.GetPublicIPAddress().String(),
				GrpcPortOnHostMachine: uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
			}
		}

		resultApiContainerHostMachineInfo = apiContainerHostMachineInfo
	}

	return resultApiContainerStatus, resultApiContainerInfo, resultApiContainerHostMachineInfo, nil
}

// Both StopEnclave and DestroyEnclave need to be able to stop enclaves, but both have a mutex guard. Because Go mutexes
//
//	aren't reentrant, DestroyEnclave can't just call StopEnclave so we use this helper function
func (manager *EnclaveManager) stopEnclaveWithoutMutex(ctx context.Context, enclaveId enclave.EnclaveUUID) error {
	_, enclaveStopErrs, err := manager.kurtosisBackend.StopEnclaves(ctx, getEnclaveByEnclaveIdFilter(enclaveId))
	if err != nil {
		return stacktrace.Propagate(err, "Attempted to stop enclave '%v' but the backend threw an error", enclaveId)
	}
	// Handle any err thrown by the backend
	enclaveStopErrStrings := []string{}
	for enclaveId, err := range enclaveStopErrs {
		if err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred stopping Enclave `%v'",
				enclaveId,
			)
			enclaveStopErrStrings = append(enclaveStopErrStrings, wrappedErr.Error())
		}
	}
	if len(enclaveStopErrStrings) > 0 {
		return stacktrace.NewError(
			"One or more errors occurred stopping the enclave(s):\n%v",
			strings.Join(
				enclaveStopErrStrings,
				"\n\n",
			))
	}
	return nil
}

func (manager *EnclaveManager) cleanEnclaves(ctx context.Context, shouldCleanAll bool) ([]string, []error, error) {
	enclaveStatusFilters := map[enclave.EnclaveStatus]bool{
		enclave.EnclaveStatus_Stopped: true,
		enclave.EnclaveStatus_Empty:   true,
	}
	if shouldCleanAll {
		enclaveStatusFilters[enclave.EnclaveStatus_Running] = true
	}

	destroyEnclaveFilters := &enclave.EnclaveFilters{
		UUIDs:    nil,
		Statuses: enclaveStatusFilters,
	}
	successfullyDestroyedEnclaves, erroredEnclaves, err := manager.kurtosisBackend.DestroyEnclaves(ctx, destroyEnclaveFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying enclaves during cleaning")
	}

	successfullyDestroyedEnclaveIdStrs := []string{}
	for enclaveId := range successfullyDestroyedEnclaves {
		successfullyDestroyedEnclaveIdStrs = append(successfullyDestroyedEnclaveIdStrs, string(enclaveId))
	}

	enclaveDestructionErrors := []error{}
	for _, destructionError := range erroredEnclaves {
		enclaveDestructionErrors = append(enclaveDestructionErrors, destructionError)
	}

	return successfullyDestroyedEnclaveIdStrs, enclaveDestructionErrors, nil
}

func (manager *EnclaveManager) getEnclavesWithoutMutex(
	ctx context.Context,
) (map[enclave.EnclaveUUID]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaves, err := manager.kurtosisBackend.GetEnclaves(ctx, getAllEnclavesFilter())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error thrown retrieving enclaves")
	}

	result := map[enclave.EnclaveUUID]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for enclaveId, enclaveObj := range enclaves {
		enclaveInfo, err := getEnclaveInfoForEnclave(ctx, manager.kurtosisBackend, enclaveObj)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting information about enclave '%v'", enclaveId)
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil

}

func (manager *EnclaveManager) Close() error {
	if err := manager.enclavePool.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the enclave pool")
	}
	return nil
}

func getEnclaveByEnclaveIdFilter(enclaveUuid enclave.EnclaveUUID) *enclave.EnclaveFilters {
	return &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUuid: true,
		},
		Statuses: nil,
	}
}

func getAllEnclavesFilter() *enclave.EnclaveFilters {
	return &enclave.EnclaveFilters{
		UUIDs:    map[enclave.EnclaveUUID]bool{},
		Statuses: nil,
	}
}

func getApiContainerByEnclaveIdFilter(enclaveId enclave.EnclaveUUID) *api_container.APIContainerFilters {
	return &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveUUID]bool{
			enclaveId: true,
		},
		Statuses: nil,
	}
}

func getEnclaveInfoForEnclave(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclave *enclave.Enclave) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaveUuid := enclave.GetUUID()
	enclaveUuidStr := string(enclaveUuid)
	apiContainerStatus, apiContainerInfo, apiContainerHostMachineInfo, err := getEnclaveApiContainerInformation(ctx, kurtosisBackend, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get information on the API container of enclave '%v', instead an error occurred.", enclaveUuid)
	}
	enclaveContainersStatus, err := getEnclaveContainersStatusFromEnclaveStatus(enclave.GetStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get EnclaveContainersStatus from the enclave status of enclave '%v', but an error occurred", enclaveUuid)
	}

	creationTimestamp, err := getEnclaveCreationTimestamp(enclave)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the creation timestamp for enclave with UUID '%v'", enclave.GetUUID())
	}

	enclaveName := enclave.GetName()

	return &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
		EnclaveUuid:                 enclaveUuidStr,
		ShortenedUuid:               uuid_generator.ShortenedUUIDString(enclaveUuidStr),
		Name:                        enclaveName,
		ContainersStatus:            enclaveContainersStatus,
		ApiContainerStatus:          apiContainerStatus,
		ApiContainerInfo:            apiContainerInfo,
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		CreationTime:                creationTimestamp,
	}, nil
}

// this should be called from a thread safe context
func (manager *EnclaveManager) getEnclaveUuidForIdentifierUnlocked(ctx context.Context, enclaveIdentifier string) (enclave.EnclaveUUID, error) {
	enclaves, err := manager.getEnclavesWithoutMutex(ctx)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting enclaves to look up if identifier '%v' is a valid uuid", enclaveIdentifier)
	}

	if _, found := enclaves[enclave.EnclaveUUID(enclaveIdentifier)]; found {
		return enclave.EnclaveUUID(enclaveIdentifier), nil
	}

	enclaveShortenedUuidEnclaveUuidMap := map[string][]enclave.EnclaveUUID{}
	enclaveNameUuidMap := map[string][]enclave.EnclaveUUID{}

	for enclaveUuid, enclave := range enclaves {
		enclaveNameUuidMap[enclave.Name] = append(enclaveNameUuidMap[enclave.Name], enclaveUuid)
		enclaveShortenedUuidEnclaveUuidMap[enclave.ShortenedUuid] = append(enclaveShortenedUuidEnclaveUuidMap[enclave.ShortenedUuid], enclaveUuid)
	}

	if matches, found := enclaveShortenedUuidEnclaveUuidMap[enclaveIdentifier]; found {
		if len(matches) == validNumberOfUuidMatches {
			return matches[0], nil
		} else if len(matches) > validNumberOfUuidMatches {
			return "", stacktrace.NewError("Found multiple enclaves '%v' matching shortened uuid '%v'. Please use a uuid to be more specific", matches, enclaveIdentifier)
		}
	}

	if matches, found := enclaveNameUuidMap[enclaveIdentifier]; found {
		if len(matches) == validNumberOfUuidMatches {
			return matches[0], nil
		} else if len(matches) > validNumberOfUuidMatches {
			return "", stacktrace.NewError("Found multiple enclaves '%v' matching name '%v'. Please use a uuid to be more specific", matches, enclaveIdentifier)
		}
	}

	return "", stacktrace.NewError("Couldn't find enclave uuid for identifier '%v'", enclaveIdentifier)
}

// Returns nil if apiContainerMap is empty
func getFirstApiContainerFromMap(apiContainerMap map[enclave.EnclaveUUID]*api_container.APIContainer) *api_container.APIContainer {
	firstApiContainerFound := (*api_container.APIContainer)(nil)
	for _, apiContainer := range apiContainerMap {
		firstApiContainerFound = apiContainer
		break
	}
	return firstApiContainerFound
}

func getEnclaveContainersStatusFromEnclaveStatus(status enclave.EnclaveStatus) (kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus, error) {
	switch status {
	case enclave.EnclaveStatus_Empty:
		return kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_EMPTY, nil
	case enclave.EnclaveStatus_Stopped:
		return kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_STOPPED, nil
	case enclave.EnclaveStatus_Running:
		return kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING, nil
	default:
		// EnclaveContainersStatus is of type int32, cannot convert nil to int32 returning -1
		return -1, stacktrace.NewError("Unrecognized enclave status '%v'; this is a bug in Kurtosis", status.String())
	}
}

func getApiContainerStatusFromContainerStatus(status container_status.ContainerStatus) (kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus, error) {
	switch status {
	case container_status.ContainerStatus_Running:
		return kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING, nil
	case container_status.ContainerStatus_Stopped:
		return kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_STOPPED, nil
	default:
		// EnclaveAPIContainerStatus is of type int32, cannot convert nil to int32 returning -1
		return -1, stacktrace.NewError("Unrecognized container status '%v'; this is a bug in Kurtosis", status.String())
	}
}

func getEnclaveCreationTimestamp(enclave *enclave.Enclave) (*timestamppb.Timestamp, error) {
	enclaveCreationTime := enclave.GetCreationTime()
	if enclaveCreationTime == nil {
		return nil, stacktrace.NewError("Expected to get the enclave creation time for enclave '%+v' but it's nil, this is a bug in Kurtosis", enclave)
	}
	enclaveCreationTimestamp := timestamppb.New(*enclaveCreationTime)

	return enclaveCreationTimestamp, nil
}

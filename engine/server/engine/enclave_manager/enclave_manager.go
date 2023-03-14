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

	apiContainerListenGrpcProxyPortNumInsideNetwork = uint16(7444)

	getRandomEnclaveIdRetries = uint16(5)

	validNumberOfUuidMatches = 1

	errorDelimiter = ", "

	enclaveNameNotFound = "Name Not Found"
)

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
}

func NewEnclaveManager(
	kurtosisBackend backend_interface.KurtosisBackend,
	apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier,
) *EnclaveManager {
	return &EnclaveManager{
		mutex:           &sync.Mutex{},
		kurtosisBackend: kurtosisBackend,
		apiContainerKurtosisBackendConfigSupplier: apiContainerKurtosisBackendConfigSupplier,
		allExistingAndHistoricalIdentifiers:       []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{},
	}
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//
//	is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// If blank, will use the default
	apiContainerImageVersionTag string,
	apiContainerLogLevel logrus.Level,
	//If blank, will use a random one
	enclaveName string,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating UUID for enclave with supplied name '%v'", enclaveName)
	}
	enclaveUuid := enclave.EnclaveUUID(uuid)

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

	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled
	// Create Enclave with kurtosisBackend
	newEnclave, err := manager.kurtosisBackend.CreateEnclave(setupCtx, enclaveUuid, enclaveName, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with name `%v` and uuid '%v'", enclaveName, enclaveUuid)
	}
	shouldDestroyEnclave := true
	defer func() {
		if shouldDestroyEnclave {
			_, destroyEnclaveErrs, err := manager.kurtosisBackend.DestroyEnclaves(teardownCtx, getEnclaveByEnclaveIdFilter(enclaveUuid))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the backend and destroy enclave '%v', but an error occurred:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
				return
			}
			for enclaveUuid, err := range destroyEnclaveErrs {
				logrus.Errorf("Expected to be able to cleanup the enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
			}
		}
	}()

	apiContainer, err := manager.launchApiContainer(setupCtx,
		apiContainerImageVersionTag,
		apiContainerLogLevel,
		enclaveUuid,
		apiContainerListenGrpcPortNumInsideNetwork,
		apiContainerListenGrpcProxyPortNumInsideNetwork,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	shouldStopApiContainer := true
	defer func() {
		if shouldStopApiContainer {
			_, destroyApiContainerErrs, err := manager.kurtosisBackend.DestroyAPIContainers(teardownCtx, getApiContainerByEnclaveIdFilter(enclaveUuid))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the API Container for enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the backend and destroy the API container for enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
				return
			}
			for enclaveUuid, err := range destroyApiContainerErrs {
				logrus.Errorf("Expected to be able to cleanup the API Container in enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
			}
		}
	}()

	var apiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	if apiContainer.GetPublicIPAddress() != nil &&
		apiContainer.GetPublicGRPCPort() != nil &&
		apiContainer.GetPublicGRPCProxyPort() != nil {

		apiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:            apiContainer.GetPublicIPAddress().String(),
			GrpcPortOnHostMachine:      uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
			GrpcProxyPortOnHostMachine: uint32(apiContainer.GetPublicGRPCProxyPort().GetNumber()),
		}
	}

	creationTimestamp := getEnclaveCreationTimestamp(newEnclave)
	newEnclaveUuid := newEnclave.GetUUID()
	newEnclaveUuidStr := string(newEnclaveUuid)
	shortenedUuid := uuid_generator.ShortenedUUIDString(newEnclaveUuidStr)

	result := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
		EnclaveUuid:        newEnclaveUuidStr,
		Name:               newEnclave.GetName(),
		ShortenedUuid:      shortenedUuid,
		ContainersStatus:   kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING,
		ApiContainerStatus: kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING,
		ApiContainerInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
			ContainerId:                "",
			IpInsideEnclave:            apiContainer.GetPrivateIPAddress().String(),
			GrpcPortInsideEnclave:      uint32(apiContainerListenGrpcPortNumInsideNetwork),
			GrpcProxyPortInsideEnclave: uint32(apiContainerListenGrpcProxyPortNumInsideNetwork),
		},
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		CreationTime:                creationTimestamp,
	}

	enclaveIdentifier := &kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers{
		EnclaveUuid:   newEnclaveUuidStr,
		Name:          enclaveName,
		ShortenedUuid: shortenedUuid,
	}
	manager.allExistingAndHistoricalIdentifiers = append(manager.allExistingAndHistoricalIdentifiers, enclaveIdentifier)

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldDestroyEnclave = false
	shouldStopApiContainer = false
	return result, nil
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

	successfullyRemovedArtifactIds, removalErrors, err := manager.cleanEnclaves(ctx, shouldCleanAll)
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

	if len(successfullyRemovedArtifactIds) > 0 {
		logrus.Infof("Successfully removed the enclaves")
		sort.Strings(successfullyRemovedArtifactIds)
		for _, successfullyRemovedEnclaveUuidStr := range successfullyRemovedArtifactIds {
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

func (manager *EnclaveManager) GetExistingAndHistoricalEnclaveIdentifiers() []*kurtosis_engine_rpc_api_bindings.EnclaveIdentifiers {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	return manager.allExistingAndHistoricalIdentifiers
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================

func (manager *EnclaveManager) getEnclaveApiContainerInformation(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
) (
	kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo,
	*kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo,
	error,
) {
	apiContainerByEnclaveIdFilter := getApiContainerByEnclaveIdFilter(enclaveId)
	enclaveApiContainers, err := manager.kurtosisBackend.GetAPIContainers(ctx, apiContainerByEnclaveIdFilter)
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
		ContainerId:                "",
		IpInsideEnclave:            apiContainer.GetPrivateIPAddress().String(),
		GrpcPortInsideEnclave:      uint32(apiContainer.GetPrivateGRPCPort().GetNumber()),
		GrpcProxyPortInsideEnclave: uint32(apiContainer.GetPrivateGRPCProxyPort().GetNumber()),
	}
	var resultApiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	if resultApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {

		var apiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
		if apiContainer.GetPublicIPAddress() != nil &&
			apiContainer.GetPublicGRPCPort() != nil &&
			apiContainer.GetPublicGRPCProxyPort() != nil {

			apiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
				IpOnHostMachine:            apiContainer.GetPublicIPAddress().String(),
				GrpcPortOnHostMachine:      uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
				GrpcProxyPortOnHostMachine: uint32(apiContainer.GetPublicGRPCProxyPort().GetNumber()),
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
		enclaveInfo, err := manager.getEnclaveInfoForEnclave(ctx, enclaveObj)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting information about enclave '%v'", enclaveId)
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil

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

func (manager *EnclaveManager) launchApiContainer(
	ctx context.Context,
	apiContainerImageVersionTag string,
	logLevel logrus.Level,
	enclaveUuid enclave.EnclaveUUID,
	grpcListenPort uint16,
	grpcProxyListenPort uint16,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {
	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		manager.kurtosisBackend,
	)
	if apiContainerImageVersionTag != "" {
		apiContainer, err := apiContainerLauncher.LaunchWithCustomVersion(
			ctx,
			apiContainerImageVersionTag,
			logLevel,
			enclaveUuid,
			grpcListenPort,
			grpcProxyListenPort,
			isPartitioningEnabled,
			metricsUserID,
			didUserAcceptSendingMetrics,
			manager.apiContainerKurtosisBackendConfigSupplier,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with custom version '%v', but an error occurred", enclaveUuid, apiContainerImageVersionTag)
		}
		return apiContainer, nil
	}
	apiContainer, err := apiContainerLauncher.LaunchWithDefaultVersion(
		ctx,
		logLevel,
		enclaveUuid,
		grpcListenPort,
		grpcProxyListenPort,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
		manager.apiContainerKurtosisBackendConfigSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with the default version, but an error occurred", enclaveUuid)
	}
	return apiContainer, nil
}

func (manager *EnclaveManager) getEnclaveInfoForEnclave(ctx context.Context, enclave *enclave.Enclave) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaveUuid := enclave.GetUUID()
	enclaveUuidStr := string(enclaveUuid)
	apiContainerStatus, apiContainerInfo, apiContainerHostMachineInfo, err := manager.getEnclaveApiContainerInformation(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get information on the API container of enclave '%v', instead an error occurred.", enclaveUuid)
	}
	enclaveContainersStatus, err := getEnclaveContainersStatusFromEnclaveStatus(enclave.GetStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get EnclaveContainersStatus from the enclave status of enclave '%v', but an error occurred", enclaveUuid)
	}

	creationTimestamp := getEnclaveCreationTimestamp(enclave)

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

// only call this from a thread safe context
func (manager *EnclaveManager) getEnclaveNameForEnclaveUuidUnlocked(enclaveUuid string) string {
	for _, identifier := range manager.allExistingAndHistoricalIdentifiers {
		if identifier.EnclaveUuid == enclaveUuid {
			return identifier.Name
		}
	}
	return ""
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

func getEnclaveCreationTimestamp(enclave *enclave.Enclave) *timestamppb.Timestamp {
	enclaveCreationTime := enclave.GetCreationTime()

	var creationTime *timestamppb.Timestamp

	//If an enclave has a nil creation time we are going to return nil also in order to check
	//TODO remove this condition after 2023-01-01 when we are sure that there is not any old enclave created without the creation time label
	//TODO after the retro-compatibility period we shouln't support the nil value and fail loudly instead
	//Handling retro-compatibility, enclaves that did not track enclave's creation time
	if enclaveCreationTime != nil {
		creationTime = timestamppb.New(*enclaveCreationTime)
	}

	return creationTime
}

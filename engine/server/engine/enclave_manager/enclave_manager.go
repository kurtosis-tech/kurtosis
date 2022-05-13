package enclave_manager

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	apiContainerListenGrpcPortNumInsideNetwork = uint16(7443)

	apiContainerListenGrpcProxyPortNumInsideNetwork = uint16(7444)

	enclavesCleaningPhaseTitle             = "enclaves"
)

// TODO Move this to the KurtosisBackend to calculate!!
// Completeness enforced via unit test
var isContainerRunningDeterminer = map[types.ContainerStatus]bool{
	types.ContainerStatus_Paused: false,
	types.ContainerStatus_Restarting: true,
	types.ContainerStatus_Running: true,
	types.ContainerStatus_Removing: false,
	types.ContainerStatus_Dead: false,
	types.ContainerStatus_Created: false,
	types.ContainerStatus_Exited: false,
}

// Unfortunately, Docker doesn't have constants for the protocols it supports declared
var objAttrsSchemaPortProtosToDockerPortProtos = map[schema.PortProtocol]string{
	schema.PortProtocol_TCP:  "tcp",
	schema.PortProtocol_SCTP: "sctp",
	schema.PortProtcol_UDP:   "udp",
}

// Manages Kurtosis enclaves, and creates new ones in response to running tasks
type EnclaveManager struct {
	// We use Docker as our backing datastore, but it has tons of race conditions so we use this mutex to ensure
	//  enclave modifications are atomic
	mutex *sync.Mutex

	kurtosisBackend backend_interface.KurtosisBackend
}

func NewEnclaveManager(
	kurtosisBackend backend_interface.KurtosisBackend,
) *EnclaveManager {
	return &EnclaveManager{
		mutex:                              &sync.Mutex{},
		kurtosisBackend:					kurtosisBackend,
	}
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) CreateEnclave(
	setupCtx context.Context,
	// If blank, will use the default
	apiContainerImageVersionTag string,
	apiContainerLogLevel logrus.Level,
	// TODO put in coreApiVersion as a param here!
	enclaveIdStr string,
	isPartitioningEnabled bool,
	shouldPublishAllPorts bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveId := enclave.EnclaveID(enclaveIdStr)
	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled
	// Check for existing enclave with Id
	foundEnclaves, err := manager.kurtosisBackend.GetEnclaves(setupCtx, getEnclaveByEnclaveIdFilter(enclaveId))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for enclaves with ID '%v'", enclaveId)
	}
	if len(foundEnclaves) > 0 {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveId)
	}

	// Create Enclave with kurtosisBackend
	createdEnclave, err := manager.kurtosisBackend.CreateEnclave(setupCtx, enclaveId, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occured creating enclave with id `%v`", enclaveId)
	}
	shouldDestroyEnclave := true
	defer func() {
		if shouldDestroyEnclave {
			_, destroyEnclaveErrs, err := manager.kurtosisBackend.DestroyEnclaves(teardownCtx, getEnclaveByEnclaveIdFilter(enclaveId))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the kackend and destroy enclave '%v', but an error occurred:\n%v", enclaveId, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveId)
				return
			}
			for enclaveId, err := range destroyEnclaveErrs {
				logrus.Errorf("Expected to be able to cleanup the enclave '%v', but an error was thrown:\n%v", enclaveId, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveId)
			}
		}
	}()

	enclaveNetworkId := createdEnclave.GetNetworkID()
	enclaveNetworkGatewayIp := createdEnclave.GetNetworkGatewayIp()

	enclaveNetworkCidr := createdEnclave.GetNetworkCIDR()

	enclaveNetworkIpAddrTracker := createdEnclave.GetNetworkIpAddrTracker()
	apiContainerPrivateIpAddr := net.IP{}
	if enclaveNetworkIpAddrTracker != nil {
		apiContainerPrivateIpAddr, err = enclaveNetworkIpAddrTracker.GetFreeIpAddr()
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to get an interal IP address for enclave '%v', but an error occurred", enclaveId)
		}
	}

	apiContainer, err := manager.launchApiContainer(setupCtx,
		apiContainerImageVersionTag,
		apiContainerLogLevel,
		enclaveId,
		enclaveNetworkId,
		enclaveNetworkCidr,
		apiContainerListenGrpcPortNumInsideNetwork,
		apiContainerListenGrpcProxyPortNumInsideNetwork,
		enclaveNetworkGatewayIp,
		apiContainerPrivateIpAddr,
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
			_, destroyApiContainerErrs, err := manager.kurtosisBackend.DestroyAPIContainers(teardownCtx, getApiContainerByEnclaveIdFilter(enclaveId))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the API Container for enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the backend and destroy the API container for enclave '%v', but an error was thrown:\n%v", enclaveId, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveId)
				return
			}
			for enclaveId, err := range destroyApiContainerErrs {
				logrus.Errorf("Expected to be able to cleanup the API Container in enclave '%v', but an error was thrown:\n%v", enclaveId, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveId)
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

	result := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
		EnclaveId:          enclaveIdStr,
		NetworkId:          enclaveNetworkId,
		NetworkCidr:        enclaveNetworkCidr,
		ContainersStatus:   kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING,
		ApiContainerStatus: kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING,
		ApiContainerInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
			IpInsideEnclave:            apiContainerPrivateIpAddr.String(),
			GrpcPortInsideEnclave:      uint32(apiContainerListenGrpcPortNumInsideNetwork),
			GrpcProxyPortInsideEnclave: uint32(apiContainerListenGrpcProxyPortNumInsideNetwork),
		},
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
	}

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldDestroyEnclave = false
	shouldStopApiContainer = false
	return result, nil
}

// It's a liiiitle weird that we return an EnclaveInfo object (which is a Protobuf object), but as of 2021-10-21 this class
//  is only used by the EngineServerService so we might as well return the object that EngineServerService wants
func (manager *EnclaveManager) GetEnclaves(
	ctx context.Context,
) (map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaves, err := manager.getEnclavesWithoutMutex(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclaves without the mutex")
	}
	// Transform map[enclave.EnclaveId]*EnclaveInfo -> map[string]*EnclaveInfo
	enclaveMapKeyedWithStrings := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for enclaveId, enclaveInfo := range enclaves {
		enclaveMapKeyedWithStrings[string(enclaveId)] = enclaveInfo
	}

	return enclaveMapKeyedWithStrings, nil
}

func (manager *EnclaveManager) StopEnclave(ctx context.Context, enclaveIdStr string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveId := enclave.EnclaveID(enclaveIdStr)
	if err := manager.stopEnclaveWithoutMutex(ctx, enclaveId); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
	}
	return nil
}

// Destroys an enclave, deleting all objects associated with it in the container engine (containers, volumes, networks, etc.)
func (manager *EnclaveManager) DestroyEnclave(ctx context.Context, enclaveIdStr string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	enclaveId := enclave.EnclaveID(enclaveIdStr)
	enclaveDestroyFilter := &enclave.EnclaveFilters{
		IDs:      map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}
	successfullyDestroyedEnclaves, erroredEnclaves, err := manager.kurtosisBackend.DestroyEnclaves(ctx, enclaveDestroyFilter)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the enclave")
	}
	if _, found := successfullyDestroyedEnclaves[enclaveId]; found {
		return nil
	}
	destructionErr, found := erroredEnclaves[enclaveId]
	if !found {
		return stacktrace.NewError("The requested enclave ID '%v' wasn't found in the successfully-destroyed enclaves map, nor in the errors map; this is a bug in Kurtosis!", enclaveId)
	}
	return destructionErr
}

func (manager *EnclaveManager) Clean(ctx context.Context, shouldCleanAll bool) (map[string]bool, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	// TODO: Refactor with kurtosis backend
	resultSuccessfullyRemovedArtifactsIds := map[string]map[string]bool{}

	// Map of cleaning_phase_title -> (successfully_destroyed_object_id, object_destruction_errors, clean_error)
	cleaningPhaseFunctions := map[string]func() ([]string, []error, error){
		enclavesCleaningPhaseTitle: func() ([]string, []error, error) {
			return manager.cleanEnclaves(ctx, shouldCleanAll)
		},
	}

	phasesWithErrors := []string{}
	for phaseTitle, cleaningFunc := range cleaningPhaseFunctions {
		logrus.Infof("Cleaning %v...", phaseTitle)
		successfullyRemovedArtifactIds, removalErrors, err := cleaningFunc()
		if err != nil {
			logrus.Errorf("Errors occurred cleaning %v:\n%v", phaseTitle, err)
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}

		if len(successfullyRemovedArtifactIds) > 0 {
			artifactIDs := map[string]bool{}
			logrus.Infof("Successfully removed the following %v:", phaseTitle)
			sort.Strings(successfullyRemovedArtifactIds)
			for _, successfulArtifactId := range successfullyRemovedArtifactIds {
				artifactIDs[successfulArtifactId] = true
				fmt.Fprintln(logrus.StandardLogger().Out, successfulArtifactId)
			}
			resultSuccessfullyRemovedArtifactsIds[phaseTitle] = artifactIDs
		}

		if len(removalErrors) > 0 {
			logrus.Errorf("Errors occurred removing the following %v:", phaseTitle)
			for _, err := range removalErrors {
				fmt.Fprintln(logrus.StandardLogger().Out, "")
				fmt.Fprintln(logrus.StandardLogger().Out, err.Error())
			}
			phasesWithErrors = append(phasesWithErrors, phaseTitle)
			continue
		}
		logrus.Infof("Successfully cleaned %v", phaseTitle)
	}

	if len(phasesWithErrors) > 0 {
		errorStr := "Errors occurred cleaning " + strings.Join(phasesWithErrors, ", ")
		return nil, stacktrace.NewError(errorStr)
	}

	return resultSuccessfullyRemovedArtifactsIds[enclavesCleaningPhaseTitle], nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================

func (manager *EnclaveManager) getEnclaveApiContainerInformation(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
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
//  aren't reentrant, DestroyEnclave can't just call StopEnclave so we use this helper function
func (manager *EnclaveManager) stopEnclaveWithoutMutex(ctx context.Context, enclaveId enclave.EnclaveID) error {
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
		enclave.EnclaveStatus_Empty: true,
	}
	if shouldCleanAll {
		enclaveStatusFilters[enclave.EnclaveStatus_Running] = true
	}

	destroyEnclaveFilters := &enclave.EnclaveFilters{
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
) (map[enclave.EnclaveID]*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaves, err := manager.kurtosisBackend.GetEnclaves(ctx, getAllEnclavesFilter())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error thrown retrieving enclaves")
	}

	result := map[enclave.EnclaveID]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for enclaveId, enclaveObj := range enclaves {
		enclaveInfo, err := manager.getEnclaveInfoForEnclave(ctx, enclaveObj)
		if err != nil {
			return nil, stacktrace.Propagate(err,"An error occurred getting information about enclave '%v'", enclaveId)
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil

}

func getEnclaveByEnclaveIdFilter(enclaveId enclave.EnclaveID) *enclave.EnclaveFilters {
	return &enclave.EnclaveFilters{
		IDs: map[enclave.EnclaveID]bool {
			enclaveId: true,
		},
	}
}

func getAllEnclavesFilter() *enclave.EnclaveFilters {
	return &enclave.EnclaveFilters{
		IDs: map[enclave.EnclaveID]bool{},
	}
}

func getApiContainerByEnclaveIdFilter(enclaveId enclave.EnclaveID) *api_container.APIContainerFilters {
	return &api_container.APIContainerFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool {
			enclaveId: true,
		},
	}
}

func (manager *EnclaveManager) launchApiContainer(
	ctx context.Context,
	apiContainerImageVersionTag string,
	logLevel logrus.Level,
	enclaveId enclave.EnclaveID,
	networkId string,
	subnetMask string,
	grpcListenPort uint16,
	grpcProxyListenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
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
			enclaveId,
			networkId,
			subnetMask,
			grpcListenPort,
			grpcProxyListenPort,
			gatewayIpAddr,
			apiContainerIpAddr,
			isPartitioningEnabled,
			metricsUserID,
			didUserAcceptSendingMetrics,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with custom version '%v', but an error occurred", enclaveId, apiContainerImageVersionTag)
		}
		return apiContainer, nil
	}
	apiContainer, err := apiContainerLauncher.LaunchWithDefaultVersion(
		ctx,
		logLevel,
		enclaveId,
		networkId,
		subnetMask,
		grpcListenPort,
		grpcProxyListenPort,
		gatewayIpAddr,
		apiContainerIpAddr,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with the default version, but an error occurred", enclaveId)
	}
	return apiContainer, nil
}

func (manager *EnclaveManager) getEnclaveInfoForEnclave(ctx context.Context, enclave *enclave.Enclave) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaveId := enclave.GetID()
	enclaveIdStr := string(enclaveId)
	apiContainerStatus, apiContainerInfo, apiContainerHostMachineInfo, err := manager.getEnclaveApiContainerInformation(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get information on the API container of enclave '%v', instead an error occurred.", enclaveId)
	}
	enclaveContainersStatus, err := getEnclaveContainersStatusFromEnclaveStatus(enclave.GetStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get EnclaveContainersStatus from the enclave status of enclave '%v', but an error occurred", enclaveId)
	}
	return &kurtosis_engine_rpc_api_bindings.EnclaveInfo {
		EnclaveId:          enclaveIdStr,
		ContainersStatus:   enclaveContainersStatus,
		ApiContainerStatus: apiContainerStatus,
		ApiContainerInfo: apiContainerInfo,
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
	}, nil
}

// Returns nil if apiContainerMap is empty
func getFirstApiContainerFromMap(apiContainerMap map[enclave.EnclaveID]*api_container.APIContainer) *api_container.APIContainer {
	firstApiContainerFound := (*api_container.APIContainer) (nil)
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


package enclave_manager

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"io/ioutil"
	"net"
	"os"
	"path"
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
	allEnclavesDirname = "enclaves"

	apiContainerListenGrpcPortNumInsideNetwork = uint16(7443)

	apiContainerListenGrpcProxyPortNumInsideNetwork = uint16(7444)

	// NOTE: It's very important that all directories created inside the engine data directory are created with 0777
	//  permissions, because:
	//  a) the engine data directory is bind-mounted on the Docker host machine
	//  b) the engine container, and pretty much every Docker container, runs as 'root'
	//  c) the Docker host machine will not be running as root
	//  d) For the host machine to be able to read & write files inside the engine data directory, it needs to be able
	//      to access the directories inside the engine data directory
	// The longterm fix to this is probably to:
	//  1) make the engine server data a Docker volume
	//  2) have the engine server expose the engine data directory to the Docker host machine via some filesystem-sharing
	//      server, like NFS or CIFS
	// This way, we preserve the host machine's ability to write to services as if they were local on the filesystem, while
	//  actually having the data live inside a Docker volume. This also sets the stage for Kurtosis-as-a-Service (where bind-mount
	//  the engine data dirpath would be impossible).
	engineDataSubdirectoryPerms = 0777

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
	// We need this so that when the engine container starts enclaves & containers, it can tell Docker where the enclave
	//  data directory is on the host machine os Docker can bind-mount it in
	engineDataDirpathOnHostMachine string

	engineDataDirpathOnEngineContainer string
}

func NewEnclaveManager(
	kurtosisBackend backend_interface.KurtosisBackend,
	engineDataDirpathOnHostMachine string,
	engineDataDirpathOnEngineContainer string,
) *EnclaveManager {
	return &EnclaveManager{
		mutex:                              &sync.Mutex{},
		kurtosisBackend:					kurtosisBackend,
		engineDataDirpathOnHostMachine:     engineDataDirpathOnHostMachine,
		engineDataDirpathOnEngineContainer: engineDataDirpathOnEngineContainer,
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
			_, destroyEnclaveErrs, err := manager.kurtosisBackend.DestroyEnclaves(teardownCtx, getEnclaveByEnclaveIdFilter(enclaveId));
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

	// TODO Handle enclave data in backend
	_, allEnclavesDirpathOnEngineContainer := manager.getAllEnclavesDirpaths()
	if err := ensureDirpathExists(allEnclavesDirpathOnEngineContainer); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring enclaves directory '%v' exists", allEnclavesDirpathOnEngineContainer)
	}
	enclaveDataDirpathOnHostMachine, enclaveDataDirpathOnEngineContainer := manager.getEnclaveDataDirpath(enclaveId)
	if _, err := os.Stat(enclaveDataDirpathOnEngineContainer); err == nil {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave data directory already exists at '%v'", enclaveId, enclaveDataDirpathOnEngineContainer)
	}
	if err := ensureDirpathExists(enclaveDataDirpathOnEngineContainer); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred ensuring enclave data directory '%v' exists", enclaveDataDirpathOnEngineContainer)
	}
	shouldDeleteEnclaveDataDir := true
	defer func() {
		if shouldDeleteEnclaveDataDir {
			if err := os.RemoveAll(enclaveDataDirpathOnEngineContainer); err != nil {
				// 'clean' will remove this dangling directories, so this is a warn rather than an error
				logrus.Warnf("Enclave creation didn't complete successfully so we tried to remove the enclave data directory '%v', but removing threw the following error; this directory will stay around until a clean is run", enclaveDataDirpathOnEngineContainer)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
			}
		}
	}()

	enclaveNetworkId := createdEnclave.GetNetworkID()
	enclaveNetworkGatewayIp := createdEnclave.GetNetworkGatewayIp()
	enclaveNetworkCidr := createdEnclave.GetNetworkCIDR()

	enclaveNetworkIpAddrTracker := createdEnclave.GetNetworkIpAddrTracker()
	apiContainerPrivateIpAddr, err := enclaveNetworkIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to get an interal IP address for enclave '%v', but an error occurred", enclaveId)
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
		enclaveDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics)

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
		ApiContainerHostMachineInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:            apiContainer.GetPublicIPAddress().String(),
			GrpcPortOnHostMachine:      uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
			GrpcProxyPortOnHostMachine: uint32(apiContainer.GetPublicGRPCProxyPort().GetNumber()),
		},
		EnclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine,
	}

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldDestroyEnclave = false
	shouldDeleteEnclaveDataDir = false
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
	successfullyDestroyedEnclaves, erroredEnclaves, err := manager.destroyEnclavesWithoutMutex(ctx, enclaveDestroyFilter)
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
	apiContainerByEnclaveIdFilter := getApiContainerByEnclaveIdFilter(enclaveId);
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
		publicGRPCPort := apiContainer.GetPublicGRPCPort()
		publicGRPCProxyPort := apiContainer.GetPublicGRPCProxyPort()
		resultApiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:            apiContainer.GetPublicIPAddress().String(),
			GrpcPortOnHostMachine:      uint32(publicGRPCPort.GetNumber()),
			GrpcProxyPortOnHostMachine: uint32(publicGRPCProxyPort.GetNumber()),
		}
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

// TODO This is copied from Kurt Core; merge these (likely into an EngineDataVolume object)!!!
func ensureDirpathExists(absoluteDirpath string) error {
	if _, statErr := os.Stat(absoluteDirpath); os.IsNotExist(statErr) {
		if mkdirErr := os.Mkdir(absoluteDirpath, engineDataSubdirectoryPerms); mkdirErr != nil {
			return stacktrace.Propagate(
				mkdirErr,
				"Directory '%v' in the engine data volume didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	// This is necessary because the os.Mkdir might not create the directory with the perms that we want due to the umask
	// Chmod is not affected by the umask, so this will guarantee we get a directory with the perms that we want
	// NOTE: This has the added benefit of, if this directory already exists (due to the user running an old engine from
	//  before 2021-11-16), we'll correct the perms
	if err := os.Chmod(absoluteDirpath, engineDataSubdirectoryPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred setting the permissions on directory '%v' to '%v'",
			absoluteDirpath,
			engineDataSubdirectoryPerms,
		)
	}
	return nil
}

func (manager *EnclaveManager) getAllEnclavesDirpaths() (onHostMachine string, onEngineContainer string) {
	onHostMachine = path.Join(
		manager.engineDataDirpathOnHostMachine,
		allEnclavesDirname,
	)
	onEngineContainer = path.Join(
		manager.engineDataDirpathOnEngineContainer,
		allEnclavesDirname,
	)
	return
}

func (manager *EnclaveManager) getEnclaveDataDirpath(enclaveId enclave.EnclaveID) (onHostMachine string, onEngineContainer string) {
	enclaveIdStr := string(enclaveId)
	allEnclavesOnHostMachine, allEnclavesOnEngineContainer := manager.getAllEnclavesDirpaths()
	onHostMachine = path.Join(
		allEnclavesOnHostMachine,
		enclaveIdStr,
	)
	onEngineContainer = path.Join(
		allEnclavesOnEngineContainer,
		enclaveIdStr,
	)
	return
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
	successfullyDestroyedEnclaves, erroredEnclaves, err := manager.destroyEnclavesWithoutMutex(ctx, destroyEnclaveFilters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying enclaves during cleaning")
	}

	// TODO: use kurtosis_backend to clean up enclave data on disk
	//remove dangling folders if any
	if err = manager.deleteDanglingDirectories(ctx); err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to delete dangling directories")
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
	for enclaveId, enclave := range enclaves {
		enclaveInfo, err := manager.getEnclaveInfoForEnclave(ctx, enclave)
		if err != nil {
			return nil, stacktrace.Propagate(err,"An error occurred getting information about enclave '%v'", enclaveId)
		}
		result[enclaveId] = enclaveInfo
	}
	return result, nil

}

// Destroys an enclave, deleting all objects associated with it in the container engine (containers, volumes, networks, etc.)
func (manager *EnclaveManager) destroyEnclavesWithoutMutex(ctx context.Context, filters *enclave.EnclaveFilters) (
	map[enclave.EnclaveID]bool,
	map[enclave.EnclaveID]error,
	error,
) {
	successfulEnclaveIds, enclaveDestroyErrs, err := manager.kurtosisBackend.DestroyEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "The backend threw an error destroying enclaves using filters: %+v", filters)
	}

	// TODO Remove this when the enclave data lives in a volume! As it currently stands, it's possible to delete an enclave
	//  WITHOUT deleting the enclave data directory!
	for enclaveId := range successfulEnclaveIds {
		// Next, remove the enclave data dir (if it exists)
		_, enclaveDataDirpathOnEngineContainer := manager.getEnclaveDataDirpath(enclaveId)
		if _, statErr := os.Stat(enclaveDataDirpathOnEngineContainer); !os.IsNotExist(statErr) {
			if removeErr := os.RemoveAll(enclaveDataDirpathOnEngineContainer); removeErr != nil {
				return nil, nil, stacktrace.Propagate(removeErr, "An error occurred removing enclave data dir '%v' on engine container", enclaveDataDirpathOnEngineContainer)
			}
		}
	}

	return successfulEnclaveIds, enclaveDestroyErrs, nil

}

// TODO Get rid of this when enclave data volumes are a thing
// Gets rid of dangling folders
func (manager *EnclaveManager) deleteDanglingDirectories(ctx context.Context) error {
	allEnclaves, err := manager.kurtosisBackend.GetEnclaves(ctx, &enclave.EnclaveFilters{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the currently-active enclaves")
	}

	_, allEnclavesDirpathOnEngineContainer := manager.getAllEnclavesDirpaths()
	fileInfos, err := ioutil.ReadDir(allEnclavesDirpathOnEngineContainer)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while reading the directory '%v' ", allEnclavesDirpathOnEngineContainer)
	}

	for _, fileInfo := range fileInfos {
		enclaveIdFromFileInfoName := enclave.EnclaveID(fileInfo.Name())
		_, found := allEnclaves[enclaveIdFromFileInfoName]
		if fileInfo.IsDir() && !found {
			folderPath := path.Join(allEnclavesDirpathOnEngineContainer, fileInfo.Name())
			if removeErr := os.RemoveAll(folderPath); removeErr != nil {
				return stacktrace.Propagate(removeErr, "An error occurred removing the data dir '%v' on engine container", folderPath)
			}
		}
	}

	return nil
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
	enclaveDataDirpathOnHostMachine string,
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
			enclaveDataDirpathOnHostMachine,
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
		enclaveDataDirpathOnHostMachine,
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
	enclaveDataDirpathOnHostMachine, _ := manager.getEnclaveDataDirpath(enclaveId)
	return &kurtosis_engine_rpc_api_bindings.EnclaveInfo {
		EnclaveId:          enclaveIdStr,
		ContainersStatus:   enclaveContainersStatus,
		ApiContainerStatus: apiContainerStatus,
		ApiContainerInfo: apiContainerInfo,
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		EnclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine,
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


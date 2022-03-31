package docker

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/repl"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/shell"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"time"
)
const (
	// TODO Change this to base 16 to be more compact??
	guidBase = 10

	KurtosisSocketEnvVar          = "KURTOSIS_API_SOCKET"
	EnclaveIdEnvVar               = "ENCLAVE_ID"
	EnclaveDataMountDirpathEnvVar = "ENCLAVE_DATA_DIR_MOUNTPOINT"

	enclaveDataDirMountpointOnReplContainer = "/kurtosis-enclave-data"

	shouldFetchStoppedContainersWhenGettingReplContainers = true
)

func (backendCore *DockerKurtosisBackend) CreateRepl(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	containerImageName string,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	stdoutFdInt int,
	bindMounts map[string]string,
)(
	*repl.Repl,
	error,
){

	replGuid := getReplGUID()

	enclaveNetwork, err := backendCore.getEnclaveNetworkByEnclaveId(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave network by enclave ID '%v'", enclaveId)
	}

	enclaveObjAttrsProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave with ID '%v'", enclaveId)
	}

	containerAttrs, err := enclaveObjAttrsProvider.ForInteractiveREPLContainer(replGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the repl container attributes for repl with GUID '%v'", replGuid)
	}
	containerName := containerAttrs.GetName()
	containerDockerLabels := containerAttrs.GetLabels()

	labels := map[string]string{}
	for dockerLabelKey, dockerLabelValue := range containerDockerLabels {
		labels[dockerLabelKey.GetString()] = dockerLabelValue.GetString()
	}

	windowSize, err := unix.IoctlGetWinsize(stdoutFdInt, unix.TIOCGWINSZ)
	if err != nil {
		return nil, stacktrace.NewError("An error occurred getting the current terminal window size")
	}
	interactiveModeTtySize := &docker_manager.InteractiveModeTtySize{
		Height: uint(windowSize.Row),
		Width:  uint(windowSize.Col),
	}

	apiContainerFilters := &api_container.APIContainerFilters{
		EnclaveIDs: map[string]bool{
			string(enclaveId): true,
		},
	}

	apiContainers, err := backendCore.GetAPIContainers(ctx, apiContainerFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting api container using filters '%+v'", apiContainerFilters)
	}
	if len(apiContainers) == 0 {
		return nil, stacktrace.NewError("No api container was found on enclave with ID '%v', it is not possible to create the repl container without this. ", enclaveId)
	}
	if len(apiContainers) > 1 {
		return nil, stacktrace.NewError("Expected to find only one api container on enclave with ID '%v', but '%v' was found; it should never happens it is a bug in Kurtosis", enclaveId, len(apiContainers))
	}

	apiContainer := apiContainers[string(enclaveId)]

	kurtosisApiContainerSocket := fmt.Sprintf("%v:%v", apiContainer.GetPrivateIPAddress(), apiContainer.GetPrivateGRPCPort())

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImageName,
		containerName.GetString(),
		enclaveNetwork.GetId(),
	).WithInteractiveModeTtySize(
		interactiveModeTtySize,
	).WithStaticIP(
		ipAddr,
	).WithEnvironmentVariables(map[string]string{
		KurtosisSocketEnvVar: kurtosisApiContainerSocket,
		EnclaveIdEnvVar: string(enclaveId),
		EnclaveDataMountDirpathEnvVar: enclaveDataDirMountpointOnReplContainer,
	}).WithBindMounts(
		bindMounts,
	).WithVolumeMounts(map[string]string{
		string(enclaveId): enclaveDataDirMountpointOnReplContainer,
	}).WithLabels(
		labels,
	).Build()

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, containerName.GetString()); err != nil {
		logrus.Warnf("Failed to pull the latest version of the repl container image '%v'; you may be running an out-of-date version", containerName.GetString())
	}

	_, _, err = backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the repl container")
	}

	newRepl := repl.NewRepl(replGuid, enclaveId)

	return newRepl, nil
}

func (backendCore *DockerKurtosisBackend) Attach(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	replGuid repl.ReplGUID,
)(
	*shell.Shell,
	error,
){

	replContainer, err := backendCore.getReplContainerByEnclaveIDAndReplGUID(ctx, enclaveId, replGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting repl container by enclave id '%v' and repl GUID '%v'", enclaveId, replGuid)
	}

	hijackedResponse, err := backendCore.dockerManager.AttachToContainer(ctx, replContainer.GetId())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't attack to the REPL container")
	}

	newShell := shell.NewShell(hijackedResponse.Conn, hijackedResponse.Reader)

	return newShell, nil
}

func (backendCore *DockerKurtosisBackend) GetRepls(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters repl.ReplFilters,
)(
	map[repl.ReplGUID]*repl.Repl,
	map[repl.ReplGUID]error,
	error,
){
	replContainers, err := backendCore.getReplContainersByEnclaveIDAndReplGUIDs(ctx, enclaveId, filters.GUIDs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting repl containers by enclave ID '%v' and repl GUIDs '%+v'", enclaveId, filters.GUIDs)
	}

	successfulRepls := map[repl.ReplGUID]*repl.Repl{}
	erroredRepls := map[repl.ReplGUID]error{}
	for guid, _ := range replContainers {
		newRepl := repl.NewRepl(guid, enclaveId)
		successfulRepls[guid] = newRepl
	}

	return successfulRepls, erroredRepls, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
func getReplGUID() repl.ReplGUID {
	now := time.Now()
	// TODO make this UnixNano to reduce risk of collisions???
	nowUnixSecs := now.Unix()
	replGuidStr :=  strconv.FormatInt(nowUnixSecs, guidBase)
	replGuid := repl.ReplGUID(replGuidStr)
	return replGuid
}

func (backendCore *DockerKurtosisBackend) getReplContainersByEnclaveIDAndReplGUIDs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	replGuids map[repl.ReplGUID]bool,
) (map[service.ServiceGUID]*types.Container, error) {


	enclaveContainers, err := backendCore.getEnclaveContainers(ctx, enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave containers for enclave with ID '%v'", enclaveId)
	}

	userServiceContainers := map[service.ServiceGUID]*types.Container{}
	for _, container := range enclaveContainers {
		if isUserServiceContainer(container) {
			for userServiceGuid := range userServiceGuids {
				if hasGuidLabel(container, string(userServiceGuid)){
					userServiceContainers[userServiceGuid] = container
				}
			}
		}
	}
	return userServiceContainers, nil
}

func (backendCore *DockerKurtosisBackend) getReplContainersByEnclaveIDAndReplGUIDs(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	replGuids map[repl.ReplGUID]bool,
) (map[repl.ReplGUID]*types.Container, error) {

	searchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString(): label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.InteractiveREPLContainerTypeLabelValue.GetString(),
	}
	foundContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, searchLabels, shouldFetchStoppedContainersWhenGettingReplContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting containers using labels '%+v'", searchLabels)
	}

	replContainers := map[repl.ReplGUID]*types.Container{}
	for _, container := range foundContainers {
		for userServiceGuid := range replGuids {
			//TODO we could improve this doing only one container iteration? or is this ok this way because is not to expensive?
			if hasEnclaveIdLabel(container, enclaveId) && hasGuidLabel(container, string(userServiceGuid)){
				replContainers[userServiceGuid] = container
			}
		}
	}
	return replContainers, nil
}

// TODO AttachToRepl

// TODO GetRepls

// TODO StopRepl

// TODO DestroyRepl

// TODO RunReplExecCommand
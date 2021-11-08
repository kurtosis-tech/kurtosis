package engine_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_server_launcher"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_api_version"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/kurtosis_engine_server_docker_api"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	containerNamePrefix = "kurtosis-engine"

	networkToStartEngineContainerIn = "bridge"

	dockerSocketFilepath = "/var/run/docker.sock"
)

// Visitor that does its best to guarantee that a Kurtosis engine is running
// If the visit method doesn't return an error, then the engine started successfully
type engineExistenceGuarantor struct {
	// Storing a context in a struct is normally bad, but this visitor is short-lived and behaves more like a function
	ctx context.Context

	// Port bindings of the maybe-started, maybe-not engine (will only be present if the engine status isn't "stopped")
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding

	dockerManager *docker_manager.DockerManager

	engineImage string

	engineAPIVersion string

	logLevel logrus.Level

	// Port bindings of the engine server that is guaranteed to be started if the visiting didn't throw an error
	// Will be nil before visiting
	postVisitingHostMachinePortBinding *nat.PortBinding
}

func newEngineExistenceGuarantor(
	ctx context.Context,
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding,
	dockerManager *docker_manager.DockerManager,
	engineImage string,
	engineAPIVersion string,
	logLevel logrus.Level,
) *engineExistenceGuarantor {
	return &engineExistenceGuarantor{
		ctx:                                    ctx,
		preVisitingMaybeHostMachinePortBinding: preVisitingMaybeHostMachinePortBinding,
		dockerManager:                          dockerManager,
		engineImage:                            engineImage,
		engineAPIVersion: engineAPIVersion,
		logLevel:                               logLevel,
		postVisitingHostMachinePortBinding:     nil,
	}
}

func (guarantor *engineExistenceGuarantor) getPostVisitingHostMachinePortBinding() *nat.PortBinding {
	return guarantor.postVisitingHostMachinePortBinding
}

// If the engine is stopped, try to start it
func (guarantor *engineExistenceGuarantor) VisitStopped() error {
	logrus.Infof(
		"No Kurtosis engine was found; attempting to start one using image '%v'...",
		guarantor.engineImage,
	)
	matchingNetworks, err := guarantor.dockerManager.GetNetworksByName(guarantor.ctx, networkToStartEngineContainerIn)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred getting networks matching the network we want to start the engine in, '%v'",
			networkToStartEngineContainerIn,
		)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 && numMatchingNetworks > 1 {
		return stacktrace.NewError(
			"Expected exactly one network matching the name of the network that we want to start the engine in, '%v', but got %v",
			networkToStartEngineContainerIn,
			numMatchingNetworks,
		)
	}
	targetNetwork := matchingNetworks[0]
	targetNetworkId := targetNetwork.GetId()

	containerStartTimeUnixSecs := time.Now().Unix()
	containerName := fmt.Sprintf(
		"%v_%v",
		containerNamePrefix,
		containerStartTimeUnixSecs,
	)
	enginePortObj, err := nat.NewPort(
		kurtosis_engine_rpc_api_consts.ListenProtocol,
		fmt.Sprintf("%v", kurtosis_engine_rpc_api_consts.ListenPort),
	)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating a port object with port num '%v' and protocol '%v' to represent the engine's port",
			kurtosis_engine_rpc_api_consts.ListenPort,
			kurtosis_engine_rpc_api_consts.ListenProtocol,
		)
	}

	engineDataDirpath, err := host_machine_directories.GetEngineDataDirpath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the engine data dirpath")
	}

	envVars, err := getEngineEnvVars(guarantor.logLevel, engineDataDirpath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the engine envvars")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		enginePortObj: docker_manager.NewManualPublishingSpec(kurtosis_engine_rpc_api_consts.ListenPort),
	}

	bindMounts := map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath: dockerSocketFilepath,
		engineDataDirpath: kurtosis_engine_server_docker_api.EngineDataDirpathOnEngineContainer,
	}
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		guarantor.engineImage,
		containerName,
		targetNetworkId,
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		engine_server_launcher.EngineContainerLabels,
	).Build()

	_, hostMachinePortBindings, err := guarantor.dockerManager.CreateAndStartContainer(guarantor.ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Kurtosis engine container")
	}
	hostMachineEnginePortBinding, found := hostMachinePortBindings[enginePortObj]
	if !found {
		return stacktrace.NewError("The Kurtosis engine server started successfully, but no host machine port binding was found")
	}
	guarantor.postVisitingHostMachinePortBinding = hostMachineEnginePortBinding
	logrus.Info("Successfully started Kurtosis engine")
	return nil
}

// We could potentially try to restart the engine ourselves here, but the case where the server isn't responding is very
//  unusual and very bad so we'd rather fail loudly
func (guarantor *engineExistenceGuarantor) VisitContainerRunningButServerNotResponding() error {
	remediationCmd := fmt.Sprintf(
		"%v %v %v && %v %v %v",
		command_str_consts.KurtosisCmdStr,
		command_str_consts.EngineCmdStr,
		command_str_consts.EngineStopCmdStr,
		command_str_consts.KurtosisCmdStr,
		command_str_consts.EngineCmdStr,
		command_str_consts.EngineStartCmdStr,
	)
	return stacktrace.NewError(
		"We couldn't guarantee that a Kurtosis engine is running because we found a running engine container whose server isn't " +
			"responding; because this is a strange state, we don't automatically try to correct the problem so you'll want to manually " +
			" restart the server by running '%v'",
		remediationCmd,
	)
}

func (guarantor *engineExistenceGuarantor) VisitRunning() error {
	guarantor.postVisitingHostMachinePortBinding = guarantor.preVisitingMaybeHostMachinePortBinding
	guarantor.checkIfEngineIsUpToDate()
	return nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
func getEngineEnvVars(logLevel logrus.Level, engineDataDirpathOnHostMachine string) (map[string]string, error) {
	// TODO replace with a constructor
	args := kurtosis_engine_server_docker_api.EngineServerArgs{
		LogLevelStr:                    logLevel.String(),
		EngineDataDirpathOnHostMachine: engineDataDirpathOnHostMachine,
	}
	serializedBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred JSON-serializing engine args '%+v'", args)
	}
	result := map[string]string{
		kurtosis_engine_server_docker_api.SerializedArgsEnvVar: string(serializedBytes),
	}
	return result, nil
}

func (guarantor *engineExistenceGuarantor) checkIfEngineIsUpToDate() {

	runningEngineSemver, cliEngineSemver, err := guarantor.getRunningAndCLIEngineVersions()
	if err != nil {
		logrus.Warn("An error occurred verifying that the running engine is on the latest version; you may be running an out-of-date engine version")
		logrus.Debugf("Getting running and CLI engine versions error: %v", err)
	}

	if runningEngineSemver.LessThan(cliEngineSemver) {
		kurtosisRestartCmd := fmt.Sprintf("%v %v %v ", command_str_consts.KurtosisCmdStr, command_str_consts.EngineCmdStr, command_str_consts.EngineRestartCmdStr)
		logrus.Warningf("The currently-running Kurtosis engine version is '%v', but the latest version is '%v'", guarantor.engineAPIVersion, kurtosis_engine_api_version.KurtosisEngineApiVersion)
		logrus.Warningf("To use the latest version, run '%v'", kurtosisRestartCmd)
	} else {
		logrus.Debugf("Currently running engine version '%v' which is up-to-date", guarantor.engineAPIVersion)
	}
	return
}

func (guarantor *engineExistenceGuarantor) getRunningAndCLIEngineVersions() (*semver.Version, *semver.Version, error) {
	runningEngineSemver, err := semver.StrictNewVersion(guarantor.engineAPIVersion)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing running engine version string '%v' to semantic version", guarantor.engineAPIVersion)
	}

	kurtosisEngineAPISemver, err := semver.StrictNewVersion(kurtosis_engine_api_version.KurtosisEngineApiVersion)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing own engine version string '%v' to semantic version", kurtosis_engine_api_version.KurtosisEngineApiVersion)
	}

	return runningEngineSemver, kurtosisEngineAPISemver, nil
}

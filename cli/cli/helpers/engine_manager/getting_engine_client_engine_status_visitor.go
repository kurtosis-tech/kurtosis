package engine_manager

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_labels_schema"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
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

	// Port bindings of the engine server that is guaranteed to be started if the visiting didn't throw an error
	// Will be nil before visiting
	postVisitingHostMachinePortBinding *nat.PortBinding
}

func newEngineExistenceGuarantor(
	ctx context.Context,
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding,
	dockerManager *docker_manager.DockerManager,
	engineImage string,
) *engineExistenceGuarantor {
	return &engineExistenceGuarantor{
		ctx:                                    ctx,
		preVisitingMaybeHostMachinePortBinding: preVisitingMaybeHostMachinePortBinding,
		dockerManager:                          dockerManager,
		engineImage:                            engineImage,
		postVisitingHostMachinePortBinding:     nil,
	}
}

func (guarantor *engineExistenceGuarantor) getPostVisitingHostMachinePortBinding() *nat.PortBinding {
	return guarantor.postVisitingHostMachinePortBinding
}

// If the engine is stopped, try to start it
func (guarantor *engineExistenceGuarantor) VisitStopped() error {
	logrus.Infof("No Kurtosis engine was found; attempting to start one using image '%v'...", guarantor.engineImage)
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

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		enginePortObj: docker_manager.NewManualPublishingSpec(kurtosis_engine_rpc_api_consts.ListenPort),
	}
	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		guarantor.engineImage,
		containerName,
		targetNetworkId,
	).WithBindMounts(map[string]string{
		// Necessary so that the engine server can interact with the Docker engine
		dockerSocketFilepath: dockerSocketFilepath,
	}).WithUsedPorts(
		usedPorts,
	).WithLabels(
		engine_labels_schema.EngineContainerLabels,
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

// Nothing to do; engine is already running
func (guarantor *engineExistenceGuarantor) VisitRunning() error {
	guarantor.postVisitingHostMachinePortBinding = guarantor.preVisitingMaybeHostMachinePortBinding
	return nil
}

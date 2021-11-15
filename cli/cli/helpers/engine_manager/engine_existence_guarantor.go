package engine_manager

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// If set to empty, then we'll use whichever default version the launcher provides
	defaultEngineImageVersionTag = ""
)

// Visitor that does its best to guarantee that a Kurtosis engine is running
// If the visit method doesn't return an error, then the engine started successfully
type engineExistenceGuarantor struct {
	// Storing a context in a struct is normally bad, but this visitor is short-lived and behaves more like a function
	ctx context.Context

	// Port bindings of the maybe-started, maybe-not engine (will only be present if the engine status isn't "stopped")
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding

	dockerManager *docker_manager.DockerManager

	objAttrsProvider schema.ObjectAttributesProvider

	engineServerLauncher *engine_server_launcher.EngineServerLauncher

	imageVersionTag string

	logLevel logrus.Level

	// If an engine is currently running, then this will contain the version that the image is running with
	// If no engine is running, this will be emptystring
	maybeCurrentlyRunningEngineVersionTag string

	// Port bindings of the engine server that is guaranteed to be started if the visiting didn't throw an error
	// Will be nil before visiting
	postVisitingHostMachinePortBinding *nat.PortBinding
}

func newEngineExistenceGuarantorWithDefaultVersion(
	ctx context.Context,
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider schema.ObjectAttributesProvider,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
) *engineExistenceGuarantor {
	return newEngineExistenceGuarantorWithCustomVersion(
		ctx,
		preVisitingMaybeHostMachinePortBinding,
		dockerManager,
		objAttrsProvider,
		defaultEngineImageVersionTag,
		logLevel,
		maybeCurrentlyRunningEngineVersionTag,
	)
}

func newEngineExistenceGuarantorWithCustomVersion(
	ctx context.Context,
	preVisitingMaybeHostMachinePortBinding *nat.PortBinding,
	dockerManager *docker_manager.DockerManager,
	objAttrsProvider schema.ObjectAttributesProvider,
	imageVersionTag string,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
) *engineExistenceGuarantor {
	return &engineExistenceGuarantor{
		ctx:                                    ctx,
		preVisitingMaybeHostMachinePortBinding: preVisitingMaybeHostMachinePortBinding,
		dockerManager:                          dockerManager,
		objAttrsProvider:                       objAttrsProvider,
		engineServerLauncher:                   engine_server_launcher.NewEngineServerLauncher(dockerManager, objAttrsProvider),
		imageVersionTag:                        imageVersionTag,
		logLevel:                               logLevel,
		maybeCurrentlyRunningEngineVersionTag:  maybeCurrentlyRunningEngineVersionTag,
		postVisitingHostMachinePortBinding:     nil,  // Will be filled in upon visiting
	}
}

func (guarantor *engineExistenceGuarantor) getPostVisitingHostMachinePortBinding() *nat.PortBinding {
	return guarantor.postVisitingHostMachinePortBinding
}

// If the engine is stopped, try to start it
func (guarantor *engineExistenceGuarantor) VisitStopped() error {
	logrus.Infof("No Kurtosis engine was found; attempting to start one...")
	engineDataDirpath, err := host_machine_directories.GetEngineDataDirpath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the engine data dirpath")
	}

	var hostMachinePortBinding *nat.PortBinding
	var engineLaunchErr error
	if guarantor.imageVersionTag == defaultEngineImageVersionTag {
		hostMachinePortBinding, engineLaunchErr = guarantor.engineServerLauncher.LaunchWithDefaultVersion(
			guarantor.ctx,
			guarantor.logLevel,
			kurtosis_context.DefaultKurtosisEngineServerPortNum,
			engineDataDirpath,
		)
	} else {
		hostMachinePortBinding, engineLaunchErr = guarantor.engineServerLauncher.LaunchWithCustomVersion(
			guarantor.ctx,
			guarantor.imageVersionTag,
			guarantor.logLevel,
			kurtosis_context.DefaultKurtosisEngineServerPortNum,
			engineDataDirpath,
		)
	}
	if engineLaunchErr != nil {
		return stacktrace.Propagate(engineLaunchErr, "An error occurred launching the engine server container")
	}

	guarantor.postVisitingHostMachinePortBinding = hostMachinePortBinding
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
func (guarantor *engineExistenceGuarantor) checkIfEngineIsUpToDate() {
	runningEngineSemver, cliEngineSemver, err := guarantor.getRunningAndCLIEngineVersions()
	if err != nil {
		logrus.Warn("An error occurred getting the running engine's version; you may be running an out-of-date engine version")
		logrus.Debugf("Getting running and CLI engine versions error: %v", err)
		return
	}

	if runningEngineSemver.LessThan(cliEngineSemver) {
		kurtosisRestartCmd := fmt.Sprintf("%v %v %v", command_str_consts.KurtosisCmdStr, command_str_consts.EngineCmdStr, command_str_consts.EngineRestartCmdStr)
		logrus.Warningf("The currently-running Kurtosis engine version is '%v', but the latest version is '%v'; to restart the engine with the latest version use '%v'", guarantor.engineAPIVersion, kurtosis_engine_api_version.KurtosisEngineApiVersion, kurtosisRestartCmd)
	} else {
		logrus.Debugf("Currently running engine version '%v' which is up-to-date", guarantor.maybeCurrentlyRunningEngineVersionTag)
	}
	return
}

func (guarantor *engineExistenceGuarantor) getRunningAndCLIEngineVersions() (*semver.Version, *semver.Version, error) {
	if guarantor.maybeCurrentlyRunningEngineVersionTag == "" {
		return nil, nil, stacktrace.NewError("Needed to report the currently-running engine's version, but it's emptystring")
	}
	runningEngineSemver, err := semver.StrictNewVersion(guarantor.maybeCurrentlyRunningEngineVersionTag)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing running engine version string '%v' to semantic version", guarantor.maybeCurrentlyRunningEngineVersionTag)
	}

	launcherEngineSemverStr := engine_server_launcher.DefaultVersion
	launcherEngineSemver, err := semver.StrictNewVersion(launcherEngineSemverStr)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing CLI's engine version string '%v' to semantic version", launcherEngineSemverStr)
	}

	return runningEngineSemver, launcherEngineSemver, nil
}

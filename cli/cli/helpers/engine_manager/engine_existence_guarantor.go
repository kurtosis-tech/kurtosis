package engine_manager

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// If set to empty, then we'll use whichever default version the launcher provides
	defaultEngineImageVersionTag = ""

	removeDeprecatedCentralizedLogsDockerCommands = "docker container rm --force kurtosis-logs-db && docker volume rm kurtosis-logs-db-vol --force && docker rm --force $(docker ps --format '{{.Names}}' | grep kurtosis-logs-collector) && docker volume rm --force $(docker volume ls --format '{{.Name}}' | grep kurtosis-logs-collector-vol)"
)

var engineRestartCmd = fmt.Sprintf(
	"%v %v %v",
	command_str_consts.KurtosisCmdStr,
	command_str_consts.EngineCmdStr,
	command_str_consts.EngineRestartCmdStr,
)

// Visitor that does its best to guarantee that a Kurtosis engine is running
// If the visit method doesn't return an error, then the engine started successfully
type engineExistenceGuarantor struct {
	// Storing a context in a struct is normally bad, but this visitor is short-lived and behaves more like a function
	ctx context.Context

	// Whether any engine that gets started should send metrics
	shouldSendMetrics bool

	// Host machine IP:port of the maybe-started, maybe-not engine (will only be present if the engine status isn't "stopped")
	preVisitingMaybeHostMachineIpAndPort *hostMachineIpAndPort

	kurtosisBackend backend_interface.KurtosisBackend

	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier

	kurtosisRemoteBackendConfigSupplier *engine_server_launcher.KurtosisRemoteBackendConfigSupplier

	engineServerLauncher *engine_server_launcher.EngineServerLauncher

	imageVersionTag string

	logLevel logrus.Level

	// If an engine is currently running, then this will contain the version that the image is running with
	// If no engine is running, this will be emptystring
	maybeCurrentlyRunningEngineVersionTag string

	// IP:port information of the engine server on the host machine that is guaranteed to be started if the visiting didn't throw an error
	// Will be empty before visiting
	postVisitingHostMachineIpAndPort *hostMachineIpAndPort

	//TODO This is a temporary hack we should remove it when centralized logs be implemented in the KubernetesBackend
	kurtosisClusterType resolved_config.KurtosisClusterType
}

func newEngineExistenceGuarantorWithDefaultVersion(
	ctx context.Context,
	preVisitingMaybeHostMachineIpAndPort *hostMachineIpAndPort,
	kurtosisBackend backend_interface.KurtosisBackend,
	shouldSendMetrics bool,
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier,
	kurtosisRemoteBackendConfigSupplier *engine_server_launcher.KurtosisRemoteBackendConfigSupplier,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
	kurtosisClusterType resolved_config.KurtosisClusterType,
) *engineExistenceGuarantor {
	return newEngineExistenceGuarantorWithCustomVersion(
		ctx,
		preVisitingMaybeHostMachineIpAndPort,
		kurtosisBackend,
		shouldSendMetrics,
		engineServerKurtosisBackendConfigSupplier,
		kurtosisRemoteBackendConfigSupplier,
		defaultEngineImageVersionTag,
		logLevel,
		maybeCurrentlyRunningEngineVersionTag,
		kurtosisClusterType,
	)
}

func newEngineExistenceGuarantorWithCustomVersion(
	ctx context.Context,
	preVisitingMaybeHostMachineIpAndPort *hostMachineIpAndPort,
	kurtosisBackend backend_interface.KurtosisBackend,
	shouldSendMetrics bool,
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier,
	kurtosisRemoteBackendConfigSupplier *engine_server_launcher.KurtosisRemoteBackendConfigSupplier,
	imageVersionTag string,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
	kurtosisClusterType resolved_config.KurtosisClusterType,
) *engineExistenceGuarantor {
	return &engineExistenceGuarantor{
		ctx:                                  ctx,
		preVisitingMaybeHostMachineIpAndPort: preVisitingMaybeHostMachineIpAndPort,
		kurtosisBackend:                      kurtosisBackend,
		engineServerKurtosisBackendConfigSupplier: engineServerKurtosisBackendConfigSupplier,
		kurtosisRemoteBackendConfigSupplier:       kurtosisRemoteBackendConfigSupplier,
		engineServerLauncher:                      engine_server_launcher.NewEngineServerLauncher(kurtosisBackend),
		imageVersionTag:                           imageVersionTag,
		logLevel:                                  logLevel,
		maybeCurrentlyRunningEngineVersionTag:     maybeCurrentlyRunningEngineVersionTag,
		postVisitingHostMachineIpAndPort:          nil, // Will be filled in upon successful visitation
		shouldSendMetrics:                         shouldSendMetrics,
		kurtosisClusterType:                       kurtosisClusterType,
	}
}

func (guarantor *engineExistenceGuarantor) getPostVisitingHostMachineIpAndPort() *hostMachineIpAndPort {
	return guarantor.postVisitingHostMachineIpAndPort
}

// If the engine is stopped, try to start it
func (guarantor *engineExistenceGuarantor) VisitStopped() error {
	logrus.Infof("No Kurtosis engine was found; attempting to start one...")

	metricsUserIdStore := metrics_user_id_store.GetMetricsUserIDStore()
	metricsUserId, err := metricsUserIdStore.GetUserID()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics user id")
	}

	var engineLaunchErr error
	if guarantor.imageVersionTag == defaultEngineImageVersionTag {
		_, _, engineLaunchErr = guarantor.engineServerLauncher.LaunchWithDefaultVersion(
			guarantor.ctx,
			guarantor.logLevel,
			kurtosis_context.DefaultGrpcEngineServerPortNum,
			kurtosis_context.DefaultGrpcProxyEngineServerPortNum,
			metricsUserId,
			guarantor.shouldSendMetrics,
			guarantor.engineServerKurtosisBackendConfigSupplier,
			guarantor.kurtosisRemoteBackendConfigSupplier,
		)
	} else {
		_, _, engineLaunchErr = guarantor.engineServerLauncher.LaunchWithCustomVersion(
			guarantor.ctx,
			guarantor.imageVersionTag,
			guarantor.logLevel,
			kurtosis_context.DefaultGrpcEngineServerPortNum,
			kurtosis_context.DefaultGrpcProxyEngineServerPortNum,
			metricsUserId,
			guarantor.shouldSendMetrics,
			guarantor.engineServerKurtosisBackendConfigSupplier,
			guarantor.kurtosisRemoteBackendConfigSupplier,
		)
	}
	if engineLaunchErr != nil {
		return stacktrace.Propagate(engineLaunchErr, "An error occurred launching the engine server container")
	}

	// TODO Replace hacky method of defaulting engine connection to localhost on predetermined port
	guarantor.postVisitingHostMachineIpAndPort = getDefaultKurtosisEngineLocalhostMachineIpAndPort()
	logrus.Info("Successfully started Kurtosis engine")
	return nil
}

// We could potentially try to restart the engine ourselves here, but the case where the server isn't responding is very
// unusual and very bad, so we'd rather fail loudly
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
		"We couldn't guarantee that a Kurtosis engine is running because we found a running engine container whose server isn't "+
			"responding; because this is a strange state, we don't automatically try to correct the problem so you'll want to manually "+
			" restart the server by running '%v'",
		remediationCmd,
	)
}

func (guarantor *engineExistenceGuarantor) VisitRunning() error {

	//TODO(centralized-logs-infra-deprecation) remove this code in the future when people don't have centralized logs infra running
	if guarantor.kurtosisClusterType == resolved_config.KurtosisClusterType_Docker {
		ctx := context.Background()
		guarantor.ensureDestroyDeprecatedCentralizedLogsResources(ctx)
	}

	guarantor.postVisitingHostMachineIpAndPort = guarantor.preVisitingMaybeHostMachineIpAndPort
	runningEngineSemver, cliEngineSemver, err := guarantor.getRunningAndCLIEngineVersions()
	if err != nil {
		logrus.Warn("An error occurred getting the running engine's version; you may be running an out-of-date engine version")
		logrus.Debugf("Getting running and CLI engine versions error: %v", err)
		return nil
	}

	cliEngineMajorVersion := cliEngineSemver.Major()
	cliEngineMinorVersion := cliEngineSemver.Minor()
	runningEngineMajorVersion := runningEngineSemver.Major()
	runningEngineMinorVersion := runningEngineSemver.Minor()
	doApiVersionsMatch := cliEngineMajorVersion == runningEngineMajorVersion && cliEngineMinorVersion == runningEngineMinorVersion
	// If the major.minor versions don't match, there's an API break that could cause the CLI to fail, so we force the user to
	//  restart their engine server
	if !doApiVersionsMatch {
		logrus.Errorf(
			"The engine server API version that the CLI expects, '%v', doesn't match the running engine server API version, '%v'; this would cause broken functionality so "+
				"you'll need to restart the engine to get the correct version by running '%v'",
			fmt.Sprintf("%v.%v", cliEngineMajorVersion, cliEngineMinorVersion),
			fmt.Sprintf("%v.%v", runningEngineMajorVersion, runningEngineMinorVersion),
			engineRestartCmd,
		)
		return stacktrace.NewError(
			"An API version mismatch was detected between the running engine version '%v' and the engine version the CLI expects, '%v'",
			runningEngineSemver.String(),
			cliEngineSemver.String(),
		)
	}

	if runningEngineSemver.Equal(cliEngineSemver) {
		return nil
	}

	if runningEngineSemver.LessThan(cliEngineSemver) {
		return stacktrace.NewError("The current version of the CLI '%s' is greater than the version of the running engine '%s' which could lead to unexpected behavior. Use '%s' to upgrade the current running engine", cliEngineSemver.String(), runningEngineSemver.String(), engineRestartCmd)
	}

	if runningEngineSemver.GreaterThan(cliEngineSemver) {
		logrus.Warningf("The version of the current running engine '%v' is greater than the version of the CLI '%v'. Please upgrade the CLI by following the steps here %v", runningEngineSemver.String(), cliEngineSemver.String(), user_support_constants.UpgradeCLIInstructionsPage)
	}
	return nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (guarantor *engineExistenceGuarantor) getRunningAndCLIEngineVersions() (*semver.Version, *semver.Version, error) {
	if guarantor.maybeCurrentlyRunningEngineVersionTag == "" {
		return nil, nil, stacktrace.NewError("Needed to report the currently-running engine's version, but it's emptystring")
	}
	runningEngineSemver, err := semver.StrictNewVersion(guarantor.maybeCurrentlyRunningEngineVersionTag)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing running engine version string '%v' to semantic version", guarantor.maybeCurrentlyRunningEngineVersionTag)
	}

	launcherEngineSemverStr := kurtosis_version.KurtosisVersion
	launcherEngineSemver, err := semver.StrictNewVersion(launcherEngineSemverStr)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing CLI's engine version string '%v' to semantic version", launcherEngineSemverStr)
	}

	return runningEngineSemver, launcherEngineSemver, nil
}

// TODO(centralized-logs-infra-deprecation) remove this code in the future when people don't have centralized logs infra running
func (guarantor *engineExistenceGuarantor) ensureDestroyDeprecatedCentralizedLogsResources(ctx context.Context) {

	// TODO(centralized-logs-resources-deprecation) remove this code in the future when people don't have any centralized logs collector and logs database running
	// we remove all centralized logs containers & volumes
	if err := guarantor.kurtosisBackend.DestroyDeprecatedCentralizedLogsResources(ctx); err != nil {
		logrus.Errorf("Attempted to remove deprecated centralized logs resources but failed with error:\n%v", err)
		logrus.Errorf("Users will have to remove the containers & volumes themselves using `%v`", removeDeprecatedCentralizedLogsDockerCommands)
	}
}

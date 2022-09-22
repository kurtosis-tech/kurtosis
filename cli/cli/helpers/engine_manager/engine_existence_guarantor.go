package engine_manager

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/metrics_user_id_store"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/engine_server_launcher"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// If set to empty, then we'll use whichever default version the launcher provides
	defaultEngineImageVersionTag = ""

	engineDataDirPermBits = 0755
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

	engineServerLauncher *engine_server_launcher.EngineServerLauncher

	imageVersionTag string

	logLevel logrus.Level

	// If an engine is currently running, then this will contain the version that the image is running with
	// If no engine is running, this will be emptystring
	maybeCurrentlyRunningEngineVersionTag string

	// IP:port information of the engine server on the host machine that is guaranteed to be started if the visiting didn't throw an error
	// Will be empty before visiting
	postVisitingHostMachineIpAndPort *hostMachineIpAndPort
}

func newEngineExistenceGuarantorWithDefaultVersion(
	ctx context.Context,
	preVisitingMaybeHostMachineIpAndPort *hostMachineIpAndPort,
	kurtosisBackend backend_interface.KurtosisBackend,
	shouldSendMetrics bool,
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
) *engineExistenceGuarantor {
	return newEngineExistenceGuarantorWithCustomVersion(
		ctx,
		preVisitingMaybeHostMachineIpAndPort,
		kurtosisBackend,
		shouldSendMetrics,
		engineServerKurtosisBackendConfigSupplier,
		defaultEngineImageVersionTag,
		logLevel,
		maybeCurrentlyRunningEngineVersionTag,
	)
}

func newEngineExistenceGuarantorWithCustomVersion(
	ctx context.Context,
	preVisitingMaybeHostMachineIpAndPort *hostMachineIpAndPort,
	kurtosisBackend backend_interface.KurtosisBackend,
	shouldSendMetrics bool,
	engineServerKurtosisBackendConfigSupplier engine_server_launcher.KurtosisBackendConfigSupplier,
	imageVersionTag string,
	logLevel logrus.Level,
	maybeCurrentlyRunningEngineVersionTag string,
) *engineExistenceGuarantor {
	return &engineExistenceGuarantor{
		ctx:                                  ctx,
		preVisitingMaybeHostMachineIpAndPort: preVisitingMaybeHostMachineIpAndPort,
		kurtosisBackend:                      kurtosisBackend,
		engineServerKurtosisBackendConfigSupplier: engineServerKurtosisBackendConfigSupplier,
		engineServerLauncher:                      engine_server_launcher.NewEngineServerLauncher(kurtosisBackend),
		imageVersionTag:                           imageVersionTag,
		logLevel:                                  logLevel,
		maybeCurrentlyRunningEngineVersionTag:     maybeCurrentlyRunningEngineVersionTag,
		postVisitingHostMachineIpAndPort:          nil, // Will be filled in upon successful visitation
		shouldSendMetrics:                         shouldSendMetrics,
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
			kurtosis_context.DefaultHttpLogsCollectorPortNum,
			metricsUserId,
			guarantor.shouldSendMetrics,
			guarantor.engineServerKurtosisBackendConfigSupplier,
		)
	} else {
		_, _, engineLaunchErr = guarantor.engineServerLauncher.LaunchWithCustomVersion(
			guarantor.ctx,
			guarantor.imageVersionTag,
			guarantor.logLevel,
			kurtosis_context.DefaultGrpcEngineServerPortNum,
			kurtosis_context.DefaultGrpcProxyEngineServerPortNum,
			kurtosis_context.DefaultHttpLogsCollectorPortNum,
			metricsUserId,
			guarantor.shouldSendMetrics,
			guarantor.engineServerKurtosisBackendConfigSupplier,
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
		"We couldn't guarantee that a Kurtosis engine is running because we found a running engine container whose server isn't "+
			"responding; because this is a strange state, we don't automatically try to correct the problem so you'll want to manually "+
			" restart the server by running '%v'",
		remediationCmd,
	)
}

func (guarantor *engineExistenceGuarantor) VisitRunning() error {
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
	// If the major.minor versions don't match, there's an API break that could cause the CLI to fail so we force the user to
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
	if runningEngineSemver.LessThan(cliEngineSemver) {
		logrus.Warningf(
			"The currently-running Kurtosis engine version is '%v' but the latest version is '%v'; you can pull the latest fixes by running '%v'",
			runningEngineSemver.String(),
			cliEngineSemver.String(),
			engineRestartCmd,
		)
	} else {
		logrus.Debugf("Currently running engine version '%v' is >= the version the CLI expects", guarantor.maybeCurrentlyRunningEngineVersionTag)
	}
	return nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
func (guarantor *engineExistenceGuarantor) getRunningAndCLIEngineVersions() (*semver.Version, *semver.Version, error) {
	if guarantor.maybeCurrentlyRunningEngineVersionTag == "" {
		return nil, nil, stacktrace.NewError("Needed to report the currently-running engine's version, but it's emptystring")
	}
	runningEngineSemver, err := semver.StrictNewVersion(guarantor.maybeCurrentlyRunningEngineVersionTag)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing running engine version string '%v' to semantic version", guarantor.maybeCurrentlyRunningEngineVersionTag)
	}

	launcherEngineSemverStr := engine_server_launcher.KurtosisEngineVersion
	launcherEngineSemver, err := semver.StrictNewVersion(launcherEngineSemverStr)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred parsing CLI's engine version string '%v' to semantic version", launcherEngineSemverStr)
	}

	return runningEngineSemver, launcherEngineSemver, nil
}

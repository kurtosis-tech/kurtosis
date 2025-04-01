package start

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultEngineVersion                   = ""
	restartEngineOnSameVersionIfAnyRunning = true
	lokiImage                              = "grafana/loki:3.4.2"
	lokiContainerName                      = "kurtosis-loki"
	lokiPort                               = 3100
	grafanaImage                           = "grafana/grafana:11.6.0"
	grafanaContainerName                   = "kurtosis-grafana"
	grafanaPort                            = 3000

	bridgeNetworkId = "bridge"
)

var GraflokiStartCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.GraflokiStartCmdStr,
	ShortDescription: "Starts a grafana/loki instance.",
	LongDescription:  "Starts a grafana/loki instance that the kurtosis engine will be configured to send logs to.",
	RunFunc:          run,
}

var emptyDockerClientOpts = []client.Opt{}
var lokiContainerLabels = map[string]string{
	"kurtosis-loki": "true",
}

func run(
	ctx context.Context,
	_ *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
	dockerManager, err := docker_manager.CreateDockerManager(emptyDockerClientOpts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the docker manager to start grafana and loki.")

	}
	lokiNatPort := nat.Port(strconv.Itoa(lokiPort) + "/tcp")
	grafanaNatPort := nat.Port(strconv.Itoa(grafanaPort) + "/tcp")

	lokiArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(lokiImage, lokiContainerName, bridgeNetworkId).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			lokiNatPort: docker_manager.NewManualPublishingSpec(lokiPort),
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode("bridge").
		WithLabels(lokiContainerLabels).
		Build()
	_, _, err = dockerManager.CreateAndStartContainer(ctx, lokiArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating '%v' container.", lokiContainerName)
	}
	logrus.Infof("Loki container started.")

	lokiContainers, err := dockerManager.GetContainersByLabels(ctx, lokiContainerLabels, false)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting loki container '%v' by labels '%v'", lokiContainerName, lokiContainerLabels)
	}
	if len(lokiContainers) == 0 {
		return stacktrace.Propagate(err, "An error occurred getting loki container '%v' by labels '%v'", lokiContainerName, lokiContainerLabels)
	}
	if len(lokiContainers) > 1 {
		return stacktrace.Propagate(err, "An error occurred getting loki container '%v' by labels '%v'", lokiContainerName, lokiContainerLabels)
	}
	lokiContainer := lokiContainers[0]

	lokiBridgeNetworkIpAddress := fmt.Sprintf("http://%v:%v", lokiContainer.GetDefaultIpAddress(), lokiPort)
	lokiHostNetworkIpAddress := fmt.Sprintf("http://localhost:%v", lokiPort)
	if err := waitForLokiReadiness(lokiHostNetworkIpAddress); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for Loki container to become ready.")
	}

	datasourceYamlContent := fmt.Sprintf(`apiVersion: 1
datasources:
  - name: %v
    type: loki
    access: proxy
    url: %v
    isDefault: true
    editable: true
`, lokiContainerName, lokiBridgeNetworkIpAddress)
	tmpFile, err := os.CreateTemp("", "grafana-datasource-*.yaml")
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating temp datasource config.")
	}
	defer tmpFile.Close()
	if _, err := tmpFile.WriteString(datasourceYamlContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing config.")
	}

	grafanaArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(grafanaImage, grafanaContainerName, bridgeNetworkId).
		WithUsedPorts(map[nat.Port]docker_manager.PortPublishSpec{
			grafanaNatPort: docker_manager.NewManualPublishingSpec(grafanaPort),
		}).
		WithEnvironmentVariables(map[string]string{
			"GF_SECURITY_ALLOW_EMBEDDING": "true",
			"GF_AUTH_ANONYMOUS_ENABLED":   "true",
			"GF_AUTH_ANONYMOUS_ORG_ROLE":  "Admin",
		}).
		WithBindMounts(map[string]string{
			tmpFile.Name(): "/etc/grafana/provisioning/datasources/loki.yaml",
		}).
		WithFetchingLatestImageIfMissing().
		WithRestartPolicy(docker_manager.RestartOnFailure).
		WithNetworkMode("bridge").
		Build()
	if _, _, err = dockerManager.CreateAndStartContainer(ctx, grafanaArgs); err != nil {
		return stacktrace.Propagate(err, "An error creating creating '%v' container.", grafanaContainerName)
	}
	logrus.Infof("Grafana container started.")

	lokiSink := map[string]map[string]interface{}{
		"loki": {
			"type":     "loki",
			"endpoint": lokiBridgeNetworkIpAddress,
			"encoding": map[string]string{
				"codec": "json",
			},
			"labels": map[string]string{
				"job": "kurtosis",
			},
		},
	}
	err = RestartEngineWithLokiSink(ctx, lokiSink)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred restarting engine to be configured to send logs to Loki.")
	}
	logrus.Infof("Configuring engine to send logs to Loki.")

	out.PrintOutLn(fmt.Sprintf("Grafana running at http://localhost:%v", grafanaPort))
	return nil
}

func RestartEngineWithLokiSink(ctx context.Context, lokiSink logs_aggregator.Sinks) error {
	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	var engineClientCloseFunc func() error
	var restartEngineErr error
	dontRestartAPIContainers := false
	_, engineClientCloseFunc, restartEngineErr = engineManager.RestartEngineIdempotently(ctx, defaults.DefaultEngineLogLevel, defaultEngineVersion, restartEngineOnSameVersionIfAnyRunning, defaults.DefaultEngineEnclavePoolSize, defaults.DefaultEnableDebugMode, defaults.DefaultGitHubAuthTokenOverride, dontRestartAPIContainers, defaults.DefaultDomain, defaults.DefaultLogRetentionPeriod, lokiSink)
	if restartEngineErr != nil {
		return stacktrace.Propagate(restartEngineErr, "An error occurred restarting the Kurtosis engine")
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()
	return nil
}

func waitForLokiReadiness(lokiHost string) error {
	const (
		readyPath   = "/ready"
		retryDelay  = 1 * time.Second
		maxAttempts = 30
	)

	url := lokiHost + readyPath

	for i := 0; i < maxAttempts; i++ {
		logrus.Infof("Requesting %v...", url)
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(retryDelay)
	}

	return stacktrace.NewError("Loki did not become ready after 30 attempts")
}

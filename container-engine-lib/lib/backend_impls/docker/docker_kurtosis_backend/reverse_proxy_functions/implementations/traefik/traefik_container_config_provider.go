package traefik

import (
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
	mkdirCmdName     = "mkdir"
	shCmdFlag        = "-c"
)

type traefikContainerConfigProvider struct {
	config *reverse_proxy.ReverseProxyConfig
}

func newTraefikContainerConfigProvider(config *reverse_proxy.ReverseProxyConfig) *traefikContainerConfigProvider {
	return &traefikContainerConfigProvider{config: config}
}

func (traefik *traefikContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	httpPort uint16,
	dashboardPort uint16,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	bindMounts := map[string]string{
		// Necessary so that the reverse proxy can interact with the Docker engine
		consts.DockerSocketFilepath: consts.DockerSocketFilepath,
	}

	traefikConfigContentStr, err := traefik.config.GetConfigFileContent(configFileTemplate)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the traefik configuration content")
	}

	// Create cmd to
	// 1. create config file in appropriate location in the traefik container
	// 2. start traefik with the config file
	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v '%v' && %v '%v' > %v && %v",
			mkdirCmdName,
			configDirpath,
			printfCmdName,
			traefikConfigContentStr,
			configFilepath,
			binaryFilepath,
		),
	}

	// Traefik should ALWAYS be running
	// Thus, instruct docker to restart the container if it exits with non-zero status code for whatever reason
	restartPolicy := docker_manager.RestartPolicy(docker_manager.RestartAlways)

	defaultWait, err := port_spec.CreateWaitWithDefaultValues()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a wait with default values")
	}

	// Publish HTTP and Dashboard entrypoint ports
	privateHttpPortSpec, err := port_spec.NewPortSpec(httpPort, port_spec.TransportProtocol_TCP, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating Traefik private http port spec object using number '%v' and protocol '%v'",
			httpPort,
			consts.EngineTransportProtocol.String(),
		)
	}
	privateHttpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateHttpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private http port spec to a Docker port")
	}
	privateDashboardPortSpec, err := port_spec.NewPortSpec(dashboardPort, port_spec.TransportProtocol_TCP, consts.HttpApplicationProtocol, defaultWait)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating Traefik private dashboard port spec object using number '%v' and protocol '%v'",
			dashboardPort,
			consts.EngineTransportProtocol.String(),
		)
	}
	privateDashboardDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateDashboardPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private dashboard port spec to a Docker port")
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateHttpDockerPort:      docker_manager.NewManualPublishingSpec(httpPort),
		privateDashboardDockerPort: docker_manager.NewManualPublishingSpec(dashboardPort),
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithEntrypointArgs(
		[]string{
			shBinaryFilepath,
		},
	).WithCmdArgs(
		overrideCmd,
	).WithRestartPolicy(
		restartPolicy,
	).Build()

	return createAndStartArgs, nil
}

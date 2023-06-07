package fluentd

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath        = "/bin/sh"
	shCmdFlag               = "-c"
	printfCmdName           = "printf"
	httpApplicationProtocol = "http"
)

type fluentdContainerConfigProvider struct {
	config *FluentdConfig
	//TODO now the httpPortNumber is configured from the client, because this will be published to the host machine until
	//TODO we productize logs search, tracked by this issue: https://github.com/kurtosis-tech/kurtosis/issues/340
	//TODO remove the httpPortNumber field when we do not publish the port again
	httpPortNumber uint16
}

func newFluentdContainerConfigProvider(config *FluentdConfig, httpPortNumber uint16) *fluentdContainerConfigProvider {
	return &fluentdContainerConfigProvider{config: config, httpPortNumber: httpPortNumber}
}

func (fluentd *fluentdContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	// TODO: potentially add a wait here to block until fluentd has started
	privateHttpPortSpec, err := port_spec.NewPortSpec(fluentd.httpPortNumber, httpTransportProtocol, httpApplicationProtocol, nil)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Loki container's private HTTP port spec object using number '%v' and protocol '%v'",
			fluentd.httpPortNumber,
			httpTransportProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (fluentd *fluentdContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	logsDatabaseVolumeName string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	privateHttpPortSpec, err := fluentd.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki container's private port spec")
	}

	privateHttpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateHttpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private HTTP port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateHttpDockerPort: docker_manager.NewManualPublishingSpec(privateHttpPortSpec.GetNumber()),
	}

	volumeMounts := map[string]string{
		logsDatabaseVolumeName: dirpath,
	}

	configContentStr, err := fluentd.getConfigContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	overrideCmd := []string{
		"-c",
		fmt.Sprintf(
			"%v '%v' > %v && %v %v %v",
			printfCmdName,
			configContentStr,
			configFilepath,
			binaryFilepath,
			configFileFlag,
			configFilepath,
		),
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithVolumeMounts(
		volumeMounts,
	).WithLabels(
		containerLabels,
	).WithUsedPorts(
		usedPorts,
	).WithEntrypointArgs([]string{
		"bin/sh",
	}).WithCmdArgs(
		overrideCmd,
	).Build()

	return createAndStartArgs, nil
}

func (fluentd *fluentdContainerConfigProvider) getConfigContent() (string, error) {
	configContentStr, err := fluentd.config.RenderConfig()
	if err != nil {
		return "", stacktrace.Propagate(err, "Error getting Fluentd raw config")
	}
	return configContentStr, nil
}

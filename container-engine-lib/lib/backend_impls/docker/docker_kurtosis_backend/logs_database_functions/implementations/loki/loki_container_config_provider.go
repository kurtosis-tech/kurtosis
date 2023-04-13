package loki

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"gopkg.in/yaml.v3"
)

const (
	shBinaryFilepath        = "/bin/sh"
	shCmdFlag               = "-c"
	printfCmdName           = "printf"
	httpApplicationProtocol = "http"
)

type lokiContainerConfigProvider struct {
	config *LokiConfig
	//TODO now the httpPortNumber is configured from the client, because this will be published to the host machine until
	//TODO we productize logs search, tracked by this issue: https://github.com/kurtosis-tech/kurtosis/issues/340
	//TODO remove the httpPortNumber field when we do not publish the port again
	httpPortNumber uint16
}

func newLokiContainerConfigProvider(config *LokiConfig, httpPortNumber uint16) *lokiContainerConfigProvider {
	return &lokiContainerConfigProvider{config: config, httpPortNumber: httpPortNumber}
}

func (loki *lokiContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(loki.httpPortNumber, httpTransportProtocol, httpApplicationProtocol, nil)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Loki container's private HTTP port spec object using number '%v' and protocol '%v'",
			loki.httpPortNumber,
			httpTransportProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (loki *lokiContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	logsDatabaseVolumeName string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	privateHttpPortSpec, err := loki.GetPrivateHttpPortSpec()
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

	logsDatabaseConfigContentStr, err := loki.getConfigContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v %v > %v && %v '%v' > %v && %v %v=%v",
			printfCmdName,
			runtimeConfigFileInitialContent,
			runtimeConfigFilepath,
			printfCmdName,
			logsDatabaseConfigContentStr,
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
	).WithEntrypointArgs(
		[]string{
			shBinaryFilepath,
		},
	).WithCmdArgs(
		overrideCmd,
	).Build()

	return createAndStartArgs, nil
}

func (loki *lokiContainerConfigProvider) getConfigContent() (string, error) {
	lokiConfigYAMLContent, err := yaml.Marshal(loki.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred marshalling Loki config '%+v'", loki.config)
	}
	lokiConfigYAMLContentStr := string(lokiConfigYAMLContent)
	return lokiConfigYAMLContentStr, nil
}

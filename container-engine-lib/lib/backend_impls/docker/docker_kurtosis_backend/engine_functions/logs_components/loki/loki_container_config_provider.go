package loki

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"gopkg.in/yaml.v3"
)

const (
	shBinaryFilepath = "/bin/sh"
	shCmdFlag        = "-c"
	printfCmdName    = "printf"
)

type LokiContainerConfigProvider struct {
	config *LokiConfig
}

func NewLokiContainerConfigProvider(config *LokiConfig) *LokiContainerConfigProvider {
	return &LokiContainerConfigProvider{config: config}
}

func (loki *LokiContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(httpPortNumber, httpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Loki container's private HTTP port spec object using number '%v' and protocol '%v'",
			httpPortNumber,
			httpPortProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (loki *LokiContainerConfigProvider) GetContainerArgs(
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
		privateHttpDockerPort: docker_manager.NewNoPublishingSpec(),
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

func (loki *LokiContainerConfigProvider) getConfigContent() (string, error) {
	lokiConfigYAMLContent, err := yaml.Marshal(loki.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred marshalling Loki config '%+v'", loki.config)
	}
	lokiConfigYAMLContentStr := string(lokiConfigYAMLContent)
	return lokiConfigYAMLContentStr, nil
}

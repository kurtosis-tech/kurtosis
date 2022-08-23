package fluentbit

import (
	"bytes"
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"text/template"
)

const (
	localhostStr    = "localhost"
	httpProtocolStr = "http"

	waitForAvailabilityInitialDelayMilliseconds = 100
	waitForAvailabilityMaxRetries               = 20
	waitForAvailabilityRetriesDelayMilliseconds = 50
)

type FluentbitContainerConfigProvider struct {
	config         *FluentbitConfig
	httpPortNumber uint16
}

func NewFluentbitContainerConfigProvider(config *FluentbitConfig, httpPortNumber uint16) *FluentbitContainerConfigProvider {
	return &FluentbitContainerConfigProvider{config: config, httpPortNumber: httpPortNumber}
}

func (fluent *FluentbitContainerConfigProvider) GetPrivateTcpPortSpec() (*port_spec.PortSpec, error) {
	privateTcpPortSpec, err := port_spec.NewPortSpec(tcpPortNumber, tcpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit server's private TCP port spec object using number '%v' and protocol '%v'",
			tcpPortNumber,
			tcpPortProtocol,
		)
	}
	return privateTcpPortSpec, nil
}

func (fluent *FluentbitContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(fluent.httpPortNumber, httpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit server's private HTTP port spec object using number '%v' and protocol '%v'",
			fluent.httpPortNumber,
			httpPortProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (fluent *FluentbitContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	volumeName string,
	networkId string,
	dockerManager *docker_manager.DockerManager,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	privateTcpPortSpec, err := fluent.GetPrivateTcpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Fluentbit server's private TCP port spec")
	}

	privateTcpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateTcpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private TCP port spec to a Docker port")
	}

	privateHttpPortSpec, err := fluent.GetPrivateHttpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Fluentbit server's private HTTP port spec")
	}

	privateHttpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateHttpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private HTTP port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateTcpDockerPort:  docker_manager.NewNoPublishingSpec(),
		privateHttpDockerPort: docker_manager.NewManualPublishingSpec(fluent.httpPortNumber),
	}

	volumeMounts := map[string]string{
		volumeName: configDirpathInContainer,
	}

	if err := fluent.runConfigurator(context.Background(), networkId, volumeMounts, dockerManager); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running the Fluenbit configurator in network ID '%v' and with volume mounts '%+v'",
			networkId,
			volumeMounts,
		)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithUsedPorts(
		usedPorts,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	return createAndStartArgs, nil
}

func (fluent *FluentbitContainerConfigProvider) getConfigFileContent() (string, error) {

	template, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Fluenbit config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	template.Execute(templateStrBuffer, fluent.config)

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}

package fluentbit

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	httpProtocolStr = "http"
)

type fluentbitContainerConfigProvider struct {
	config         *FluentbitConfig
	tcpPortNumber uint16
	httpPortNumber uint16
}

func newFluentbitContainerConfigProvider(config *FluentbitConfig, tcpPortNumber uint16, httpPortNumber uint16) *fluentbitContainerConfigProvider {
	return &fluentbitContainerConfigProvider{config: config, tcpPortNumber:tcpPortNumber, httpPortNumber: httpPortNumber}
}

func (fluent *fluentbitContainerConfigProvider) GetPrivateTcpPortSpec() (*port_spec.PortSpec, error) {
	privateTcpPortSpec, err := port_spec.NewPortSpec(fluent.tcpPortNumber, tcpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit server's private TCP port spec object using number '%v' and protocol '%v'",
			fluent.tcpPortNumber,
			tcpPortProtocol,
		)
	}
	return privateTcpPortSpec, nil
}

func (fluent *fluentbitContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
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

func (fluent *fluentbitContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	volumeName string,
	networkId string,
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
		privateTcpDockerPort:  docker_manager.NewManualPublishingSpec(fluent.tcpPortNumber),
		privateHttpDockerPort: docker_manager.NewManualPublishingSpec(fluent.httpPortNumber),
	}

	volumeMounts := map[string]string{
		volumeName: configDirpathInContainer,
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

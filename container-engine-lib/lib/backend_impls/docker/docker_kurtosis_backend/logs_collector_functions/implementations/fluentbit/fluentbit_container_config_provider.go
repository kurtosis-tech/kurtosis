package fluentbit

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	httpProtocolStr = "http"
)

type fluentbitContainerConfigProvider struct {
	config         *FluentbitConfig
	tcpPortNumber  uint16
	httpPortNumber uint16
}

func newFluentbitContainerConfigProvider(config *FluentbitConfig, tcpPortNumber uint16, httpPortNumber uint16) *fluentbitContainerConfigProvider {
	return &fluentbitContainerConfigProvider{config: config, tcpPortNumber: tcpPortNumber, httpPortNumber: httpPortNumber}
}

func (fluent *fluentbitContainerConfigProvider) GetPrivateTcpPortSpec() (*port_spec.PortSpec, error) {
	privateTcpPortSpec, err := port_spec.NewPortSpec(fluent.tcpPortNumber, tcpTransportProtocol, httpProtocolStr, nil, "")
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit server's private TCP port spec object using number '%v' and protocol '%v'",
			fluent.tcpPortNumber,
			tcpTransportProtocol,
		)
	}
	return privateTcpPortSpec, nil
}

func (fluent *fluentbitContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(fluent.httpPortNumber, httpTransportProtocol, httpProtocolStr, nil, "")
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit server's private HTTP port spec object using number '%v' and protocol '%v'",
			fluent.httpPortNumber,
			httpTransportProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (fluent *fluentbitContainerConfigProvider) GetContainerArgs(containerName string, containerLabels map[string]string, volumeName string, networkId string) (*docker_manager.CreateAndStartContainerArgs, error) {

	volumeMounts := map[string]string{
		volumeName: configDirpathInContainer,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	return createAndStartArgs, nil
}

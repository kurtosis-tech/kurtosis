package fluentbit

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath = "/bin/sh"
	shCmdFlag        = "-c"
	printfCmdName    = "printf"
)

type Fluentbit struct {
	config *Config
}

func NewFluentbit(config *Config) *Fluentbit {
	return &Fluentbit{config: config}
}

func (fluent *Fluentbit) GetPrivateTcpPortSpec() (*port_spec.PortSpec, error) {
	privateTcpPortSpec, err := port_spec.NewPortSpec(tcpPortNumber, tcpPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Fluentbit's private TCP port spec object using number '%v' and protocol '%v'",
			tcpPortNumber,
			tcpPortProtocol,
		)
	}
	return privateTcpPortSpec, nil
}

func (fluent *Fluentbit) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	privateTcpPortSpec, err := fluent.GetPrivateTcpPortSpec()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Fluentbit's private TCP port spec")
	}

	privateTcpDockerPort, err := shared_helpers.TransformPortSpecToDockerPort(privateTcpPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private TCP port spec to a Docker port")
	}

	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateTcpDockerPort: docker_manager.NewManualPublishingSpec(tcpPortNumber),
	}

	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v '%v' > %v  && %v %v=%v",
			printfCmdName,
			configContentTemplate,
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
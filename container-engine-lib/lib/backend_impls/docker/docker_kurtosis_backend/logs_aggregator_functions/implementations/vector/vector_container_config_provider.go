package vector

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
	shCmdFlag        = "-c"
	validateCmdName  = "validate"
	httpProtocolStr  = "http"
)

type vectorContainerConfigProvider struct {
	httpPortNumber uint16
}

func newVectorContainerConfigProvider(httpPortNumber uint16) *vectorContainerConfigProvider {
	return &vectorContainerConfigProvider{
		httpPortNumber: httpPortNumber,
	}
}

func (vector *vectorContainerConfigProvider) GetPrivateHttpPortSpec() (*port_spec.PortSpec, error) {
	privateHttpPortSpec, err := port_spec.NewPortSpec(vector.httpPortNumber, httpTransportProtocol, httpProtocolStr, nil, "")
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the Vector server's private HTTP port spec object using number '%v' and protocol '%v'",
			vector.httpPortNumber,
			httpTransportProtocol,
		)
	}
	return privateHttpPortSpec, nil
}

func (vector *vectorContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	networkId string,
	logsAggregatorVolumeName string,
	logsStorageVolumeName string,
) (*docker_manager.CreateAndStartContainerArgs, error) {
	volumeMounts := map[string]string{
		logsStorageVolumeName:    logsStorageDirpath,
		logsAggregatorVolumeName: configDirpath,
	}

	// Create cmd to
	// 1. create config file in appropriate location in logs aggregator container
	// 2. start the logs aggregator with the config file
	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v %v=%v",
			binaryFilepath,
			configFileFlag,
			configFilepath,
		),
	}

	// The logs aggregator should ALWAYS be running to ensure that no logs are lost for services in enclaves
	// Thus, instruct docker to restart the container if it exits with non-zero status code for whatever reason
	restartPolicy := docker_manager.RestartPolicy(docker_manager.RestartAlways)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithVolumeMounts(
		volumeMounts,
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

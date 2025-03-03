package vector

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
	shCmdFlag        = "-c"
	validateCmdName  = "validate"
)

type vectorContainerConfigProvider struct {
	config *VectorConfig
}

func newVectorContainerConfigProvider(config *VectorConfig) *vectorContainerConfigProvider {
	return &vectorContainerConfigProvider{config: config}
}

func (vector *vectorContainerConfigProvider) GetInitContainerArgs(
	containerName string,
	containerLabels map[string]string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {
	logsAggregatorConfigContentStr, err := vector.config.getConfigFileContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	// Create cmd to
	// 1. create config file in appropriate location in logs aggregator container
	// 2. start the logs aggregator with the config file
	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v '%v' > %v && %v %v %v",
			printfCmdName,
			logsAggregatorConfigContentStr,
			configFilepath,
			binaryFilepath,
			validateCmdName,
			configFilepath,
		),
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithEntrypointArgs(
		[]string{
			shBinaryFilepath,
		},
	).WithCmdArgs(
		overrideCmd,
	).WithSkipSuccessfulStartCheck(
		true,
	).Build()

	return createAndStartArgs, nil
}

func (vector *vectorContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	networkId string,
	logsStorageVolumeName string,
) (*docker_manager.CreateAndStartContainerArgs, error) {
	volumeMounts := map[string]string{
		logsStorageVolumeName: logsStorageDirpath,
	}

	logsAggregatorConfigContentStr, err := vector.config.getConfigFileContent()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Loki server's configuration content")
	}

	// Create cmd to
	// 1. create config file in appropriate location in logs aggregator container
	// 2. start the logs aggregator with the config file
	overrideCmd := []string{
		shCmdFlag,
		fmt.Sprintf(
			"%v '%v' > %v && %v %v=%v",
			printfCmdName,
			logsAggregatorConfigContentStr,
			configFilepath,
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

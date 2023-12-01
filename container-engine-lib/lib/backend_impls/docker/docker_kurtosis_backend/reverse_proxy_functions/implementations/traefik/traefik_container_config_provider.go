package traefik

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	shBinaryFilepath = "/bin/sh"
	printfCmdName    = "printf"
	mkdirCmdName     = "mkdir"
	shCmdFlag        = "-c"
)

type traefikContainerConfigProvider struct {
	config *TraefikConfig
}

func newTraefikContainerConfigProvider(config *TraefikConfig) *traefikContainerConfigProvider {
	return &traefikContainerConfigProvider{config: config}
}

func (traefik *traefikContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	traefikConfigContentStr, err := traefik.config.getConfigFileContent()
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
	).WithRestartPolicy(
		restartPolicy,
	).Build()

	return createAndStartArgs, nil
}

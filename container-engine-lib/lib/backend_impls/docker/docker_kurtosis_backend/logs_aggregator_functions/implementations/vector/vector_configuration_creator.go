package vector

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	configuratorContainerNamePrefix     = "kurtosis-vector-configurator"
	configFileCreationCmdMaxRetries     = 2
	configFileCreationCmdDelayInRetries = 200 * time.Millisecond
	validationFailedExitCode            = 78
	configFileCreationSuccessExitCode   = 0
	sleepSeconds                        = 1800
)

type vectorConfigurationCreator struct {
	config *VectorConfig
}

func newVectorConfigurationCreator(config *VectorConfig) *vectorConfigurationCreator {
	return &vectorConfigurationCreator{config: config}
}

func (vector *vectorConfigurationCreator) CreateConfiguration(
	ctx context.Context,
	targetNetworkId string,
	volumeName string,
	dockerManager *docker_manager.DockerManager,
) error {

	entrypointArgs := []string{
		shBinaryFilepath,
		shCmdFlag,
		fmt.Sprintf("sleep %v", sleepSeconds),
	}

	volumeMounts := map[string]string{
		volumeName: configDirpath,
	}

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred generating a UUID for the configurator container name")
	}

	containerName := fmt.Sprintf("%s-%s", configuratorContainerNamePrefix, uuid)

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		targetNetworkId,
	).WithEntrypointArgs(
		entrypointArgs,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the logs aggregator configurator container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if err = dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the logs aggregator configurator container with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the container with ID '%v'!!!!!!", containerId)
		}
	}()

	if err := vector.createVectorConfigFileInVolume(
		ctx,
		dockerManager,
		containerId,
		configFileCreationCmdMaxRetries,
		configFileCreationCmdDelayInRetries,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the logs aggregator config file into the volume")
	}

	return nil
}

func (vector *vectorConfigurationCreator) createVectorConfigFileInVolume(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	containerId string,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {

	configFileContentStr, err := vector.getConfigFileContent()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting logs aggregator config file content")
	}

	commandStr := fmt.Sprintf(
		"%v '%v' > %v && %v %v %v",
		printfCmdName,
		configFileContentStr,
		configFilepath,
		binaryFilepath,
		validateCmdName,
		configFilepath,
	)

	execCmd := []string{
		shBinaryFilepath,
		shCmdFlag,
		commandStr,
	}
	for i := uint(0); i < maxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunUserServiceExecCommands(ctx, containerId, "", execCmd, outputBuffer)
		if err == nil {
			if exitCode == configFileCreationSuccessExitCode {
				logrus.Debugf("The logs aggregator config file with content '%v' was successfully added into the volume", configFileContentStr)
				return nil
			}

			// Vector returns a specific exit code if the validation of configurations failed
			// https://vector.dev/docs/administration/validating/#how-validation-works
			if exitCode == validationFailedExitCode {
				return stacktrace.NewError("The configuration provided to the logs aggregator component was invalid; errors are below:\n%s", outputBuffer.String())
			}

			logrus.Debugf(
				"Logs aggregator config file creation command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				configFileCreationSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Logs aggregator config file creation command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxRetries {
			time.Sleep(timeBetweenRetries)
		}
	}

	return stacktrace.NewError(
		"The logs aggregator config file creation didn't return success (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}

func (vector *vectorConfigurationCreator) getConfigFileContent() (string, error) {
	yamlBytes, err := yaml.Marshal(vector.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error marshalling config into YAML.")
	}

	return string(yamlBytes), nil
}

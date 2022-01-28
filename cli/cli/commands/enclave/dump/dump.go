package dump

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/object-attributes-schema-lib/forever_constants"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const (
	enclaveIdArg     = "enclave-id"
	outputDirpathArg = "output-dirpath"

	shouldGetStoppedContainers = true

	containerInspectResultFilename = "spec.json"
	containerLogsFilename          = "output.log"

	numContainersToDumpAtOnce = 20

	// Permisssions for the files & directories we create as a result of the dump
	createdDirPerms  = 0755
	createdFilePerms = 0644

	shouldCaptureContainerStderr = true
	shouldCaptureContainerStdout = true
	shouldFollowContainerLogs    = false

	containerSpecJsonSerializationIndent = "  "
	containerSpecJsonSerializationPrefix = ""
)

var positionalArgs = []string{
	enclaveIdArg,
	outputDirpathArg,
}

var EnclaveDumpCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveDumpCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Dumps all information about the given enclave into the given output dirpath",
	RunE:                  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIdArg]
	enclaveOutputDirpath := parsedPositionalArgs[outputDirpathArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to display the state for enclave '%v'", enclaveId)
	}

	if _, found := getEnclavesResp.EnclaveInfo[enclaveId]; !found {
		return stacktrace.NewError("No enclave with ID '%v' exists", enclaveId)
	}

	// TODO REPLACE THIS WITH A CALL TO THE ENGINE SERVER TO GET THE CONTAINER IDS!!!
	enclaveContainerSearchLabels := map[string]string{
		forever_constants.AppIDLabel:   forever_constants.AppIDValue,
		schema.EnclaveIDContainerLabel: enclaveId,
	}

	enclaveContainers, err := dockerManager.GetContainersByLabels(ctx, enclaveContainerSearchLabels, shouldGetStoppedContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the containers in enclave '%v' so their logs could be dumped to disk")
	}

	// Create output directory
	if _, err := os.Stat(enclaveOutputDirpath); !os.IsNotExist(err) {
		return stacktrace.NewError("Cannot create output directory at '%v'; directory already exists", enclaveOutputDirpath)
	}
	if err := os.Mkdir(enclaveOutputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating output directory at '%v'")
	}

	workerPool := workerpool.New(numContainersToDumpAtOnce)
	resultErrsChan := make(chan error, len(enclaveContainers))
	for _, container := range enclaveContainers {
		containerName := container.GetName()
		containerId := container.GetId()
		logrus.Debugf("Submitting job to dump info about container with name '%v' and ID '%v'", containerName, containerId)
		workerPool.Submit(func() {
			if err := dumpContainerInfo(ctx, dockerClient, enclaveOutputDirpath, containerName, containerId); err != nil {
				resultErrsChan <- stacktrace.Propagate(
					err,
					"An error occurred dumping container info for container with name '%v' and ID '%v'",
					container.GetName(),
					container.GetId(),
				)
			}
		})
	}
	workerPool.StopWait()
	close(resultErrsChan)

	allResultErrStrs := []string{}
	for resultErr := range resultErrsChan {
		allResultErrStrs = append(allResultErrStrs, resultErr.Error())
	}

	if len(allResultErrStrs) > 0 {
		allIndexedResultErrStrs := []string{}
		for idx, resultErrStr := range allResultErrStrs {
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, resultErrStr)
			allIndexedResultErrStrs = append(allIndexedResultErrStrs, indexedResultErrStr)
		}

		// NOTE: We don't use stacktrace here because the actual stacktraces we care about are the ones from the threads!
		return errors.New(fmt.Sprintf(
			"The following errors occurred when trying to dump information about enclave '%v':\n%v",
			enclaveId,
			strings.Join(allIndexedResultErrStrs, "\n\n"),
		))
	}

	logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveId, enclaveOutputDirpath)
	return nil
}

func dumpContainerInfo(
	ctx context.Context,
	dockerClient *client.Client,
	enclaveOutputDirpath string,
	containerName string,
	containerId string,
) error {
	// Make output directory
	containerOutputDirpath := path.Join(enclaveOutputDirpath, containerName)
	if err := os.Mkdir(containerOutputDirpath, createdDirPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating directory '%v' to hold the output of container with name '%v' and ID '%v'",
			containerOutputDirpath,
			containerName,
			containerId,
		)
	}

	// Write container inspect results to file
	inspectResult, err := dockerClient.ContainerInspect(
		ctx,
		containerId,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting container with ID '%v'")
	}
	jsonSerializedInspectResultBytes, err := json.MarshalIndent(inspectResult, containerSpecJsonSerializationPrefix, containerSpecJsonSerializationIndent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing the results of inspecting container with ID '%v' to JSON", containerId)
	}
	specOutputFilepath := path.Join(containerOutputDirpath, containerInspectResultFilename)
	ioutil.WriteFile(specOutputFilepath, jsonSerializedInspectResultBytes, createdFilePerms)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating file '%v' to hold the spec of container with name '%v' and ID '%v'",
			specOutputFilepath,
			containerName,
			containerId,
		)
	}

	// Write container logs to file
	containerLogsReadCloser, err := dockerClient.ContainerLogs(
		ctx,
		containerId,
		types.ContainerLogsOptions{
			ShowStderr: shouldCaptureContainerStderr,
			ShowStdout: shouldCaptureContainerStdout,
			Follow:     shouldFollowContainerLogs,
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the logs for container with ID '%v'")
	}
	defer containerLogsReadCloser.Close()
	logsOutputFilepath := path.Join(containerOutputDirpath, containerLogsFilename)
	logsOutputFp, err := os.Create(logsOutputFilepath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating file '%v' to hold the logs of container with name '%v' and ID '%v'",
			logsOutputFilepath,
			containerName,
			containerId,
		)
	}
	if _, err := stdcopy.StdCopy(logsOutputFp, logsOutputFp, containerLogsReadCloser); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the Docker container logs stream for container with name '%v' and ID '%v' to file '%v'",
			containerName,
			containerId,
			logsOutputFilepath,
		)
	}

	return nil
}

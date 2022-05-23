/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/files_artifacts_expander/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	successExitCode = 0
	failureExitCode = 1
	maxWorkers = 4

	//Files permissions for temporary file storing files artifact data: readable by all the user groups, but writable by user only
	filesArtifactTemporaryFilePermissions = 0644
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	err := runMain()
	if err != nil {
		logrus.Errorf("An error occurred when running the main function:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}

func runMain() error {
	filesArtifactExpanderArgs, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to get arguments from environment, instead a non-nil error was returned")
	}
	// Connect to the API container described in the args
	ipAddrString := filesArtifactExpanderArgs.APIContainerIpAddress
	apiContainerIpAddr := net.ParseIP(ipAddrString)
	if apiContainerIpAddr == nil {
		return stacktrace.NewError("Expected to be able parse a valid api container IP address from string '%v', instead parsed a nil IP address", ipAddrString)
	}
	apiContainerPortNum := filesArtifactExpanderArgs.ApiContainerPort
	grpcUrl := fmt.Sprintf("%v:%v", apiContainerIpAddr, apiContainerPortNum)
	apiContainerConnection, err := grpc.Dial(grpcUrl, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to create a client connection to API container at address '%v', instead a non-nil error was returned", grpcUrl)
	}
	defer apiContainerConnection.Close()
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(apiContainerConnection)
	backgroundContext := context.Background()
	// Download and extract the file artifacts in the args
	filesArtifactWorkerPool := workerpool.New(maxWorkers)
	resultErrsChan := make(chan error, len(filesArtifactExpanderArgs.FilesArtifactExpansions))
	for _, filesArtifactExpansion := range filesArtifactExpanderArgs.FilesArtifactExpansions {
		jobToSubmit := createExpandFilesArtifactJob(backgroundContext, apiContainerClient, resultErrsChan, filesArtifactExpansion)
		filesArtifactWorkerPool.Submit(jobToSubmit)
	}
	filesArtifactWorkerPool.StopWait()
	close(resultErrsChan)

	allResultErrStrs := []string{}
	for resultErr := range(resultErrsChan) {
		if resultErr != nil {
			allResultErrStrs = append(allResultErrStrs, resultErr.Error())
		}
	}

	if len(allResultErrStrs) > 0 {
		allIndexedResultErrStrs := []string{}
		for idx, resultErrStr := range allResultErrStrs {
			indexedResultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR %v <<<<<<<<<<<<<<<<<\n%v", idx, resultErrStr)
			allIndexedResultErrStrs = append(allIndexedResultErrStrs, indexedResultErrStr)
		}

		// NOTE: We don't use stacktrace here because the actual stacktraces we care about are the ones from the threads!
		return errors.New(fmt.Sprintf(
			"The following errors occurred when trying to expand files artifacts:\n%v",
			strings.Join(allIndexedResultErrStrs, "\n\n"),
		))
	}
	return nil
}

func createExpandFilesArtifactJob(ctx context.Context, apiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, resultErrsChan chan error, filesArtifactExpansion args.FilesArtifactExpansion) func() {
	return func() {
		if err := expandFilesArtifact(ctx, apiContainerClient, filesArtifactExpansion); err != nil {
			resultErrsChan <- stacktrace.Propagate(err, "An error occured expanding files artifact '%v' into directory '%v'", filesArtifactExpansion.FilesArtifactId, filesArtifactExpansion.DirPathToExpandTo)
		}
	}
}

func expandFilesArtifact(ctx context.Context, apiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, filesArtifactExpansion args.FilesArtifactExpansion) error {
	artifactId := filesArtifactExpansion.FilesArtifactId
	// Get the raw bytes of the file artifact
	downloadRequestArgs := &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs{
		Id: artifactId,
	}
	response, err := apiContainerClient.DownloadFilesArtifact(ctx, downloadRequestArgs)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to download files artifacts for files artifact with id '%v' from Kurtosis, instead a non-nil error was returned", artifactId)
	}
	// Save the bytes to file, might not be necssary if we can pipe the artifact bytes to stdin
	filesArtifactFile, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to create a temporary file for the files artifact bytes, instead a non-nil error was returned")
	}
	if err := filesArtifactFile.Close(); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to close the temporary file '%v' we created to store the downloaded files artifact, instead a non-nil error was returned", filesArtifactFile.Name())
	}
	filesArtifactFileName := filesArtifactFile.Name()
	if err := os.WriteFile(filesArtifactFileName, response.Data, filesArtifactTemporaryFilePermissions); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to save files artifact to disk at path '%v', instead a non nil error was returned", filesArtifactFileName)
	}
	// Extract the tarball to the specified location
	extractTarballCmd := exec.Command("tar", "-xzf", filesArtifactFileName, "-C", filesArtifactExpansion.DirPathToExpandTo)
	if err := extractTarballCmd.Start(); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to extract the tarball containing the files artifacts, instead a non nil error was returned")
	}
	return nil
}

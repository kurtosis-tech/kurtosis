/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"context"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/files_artifacts_expander/args"
	"github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang/grpc_file_streaming"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	successExitCode = 0
	failureExitCode = 1
	maxWorkers      = 4

	//Files permissions for temporary file storing files artifact data: readable by all the user groups, but writable by user only
	filesArtifactTemporaryFilePermissions = 0644

	forceColors   = true
	fullTimestamp = true
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               forceColors,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             fullTimestamp,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier:          nil,
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
	apiContainerConnection, err := grpc.Dial(grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	for resultErr := range resultErrsChan {
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
		return fmt.Errorf(
			"The following errors occurred when trying to expand files artifacts:\n%v",
			strings.Join(allIndexedResultErrStrs, "\n\n"),
		)
	}
	return nil
}

func createExpandFilesArtifactJob(ctx context.Context, apiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, resultErrsChan chan error, filesArtifactExpansion args.FilesArtifactExpansion) func() {
	return func() {
		if err := expandFilesArtifact(ctx, apiContainerClient, filesArtifactExpansion); err != nil {
			resultErrsChan <- stacktrace.Propagate(err, "An error occurred expanding files artifact '%v' into directory '%v'", filesArtifactExpansion.FilesIdentifier, filesArtifactExpansion.DirPathToExpandTo)
		}
	}
}

func expandFilesArtifact(ctx context.Context, apiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, filesArtifactExpansion args.FilesArtifactExpansion) error {
	artifactIdentifier := filesArtifactExpansion.FilesIdentifier
	// Get the raw bytes of the file artifact
	downloadRequestArgs := &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs{
		Identifier: artifactIdentifier,
	}
	client, err := apiContainerClient.DownloadFilesArtifact(ctx, downloadRequestArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred initiating the download of files artifact '%v'", artifactIdentifier)
	}
	clientStream := grpc_file_streaming.NewClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, any](client)
	fileContent, err := clientStream.ReceiveData(
		artifactIdentifier,
		func(dataChunk *kurtosis_core_rpc_api_bindings.StreamedDataChunk) ([]byte, string, error) {
			return dataChunk.Data, dataChunk.PreviousChunkHash, nil
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to download files artifacts for files artifact with identifier '%v' from Kurtosis, instead a non-nil error was returned", artifactIdentifier)
	}

	// Save the bytes to file, might not be necessary if we can pipe the artifact bytes to stdin
	filesArtifactFile, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return stacktrace.Propagate(err, "Expected to be able to create a temporary file for the files artifact bytes, instead a non-nil error was returned")
	}
	if err := filesArtifactFile.Close(); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to close the temporary file '%v' we created to store the downloaded files artifact, instead a non-nil error was returned", filesArtifactFile.Name())
	}

	filesArtifactFileName := filesArtifactFile.Name()
	if err := os.WriteFile(filesArtifactFileName, fileContent, filesArtifactTemporaryFilePermissions); err != nil {
		return stacktrace.Propagate(err, "Expected to be able to save files artifact to disk at path '%v', instead a non nil error was returned", filesArtifactFileName)
	}
	// Extract the tarball to the specified location
	extractTarballCmd := exec.Command("tar", "-xzf", filesArtifactFileName, "-C", filesArtifactExpansion.DirPathToExpandTo)
	if err := extractTarballCmd.Run(); err != nil {
		// Per the docs, we can downcast like so
		castedErr, ok := err.(*exec.ExitError)
		if !ok {
			return stacktrace.Propagate(err, "Command '%v' failed with an unrecognized error", extractTarballCmd.String())
		}
		return stacktrace.NewError("Command '%v' exited with an error and the following STDERR:\n%v", extractTarballCmd.String(), string(castedErr.Stderr))
	}
	return nil
}

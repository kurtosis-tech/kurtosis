/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_metadata_serialization

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
	"sync"
	"time"
)

const (
	// Time after which registration that a testsuite must call metadata serialization else an error will be thrown
	serializationTimeout = 10 * time.Second
)

type suiteMetadataSerializationService struct {
	// Mutex guarding hasSerializeBeenCalled
	mutex                                 *sync.Mutex
	hasSerializeBeenCalled                bool
	serializedSuiteMetadataOutputFilepath string
	shutdownChan chan int
}

func newSuiteMetadataSerializationService(
		serializedSuiteMetadataOutputFilepath string,
		shutdownChan chan int) *suiteMetadataSerializationService {
	return &suiteMetadataSerializationService{
		mutex: &sync.Mutex{},
		hasSerializeBeenCalled: false,
		serializedSuiteMetadataOutputFilepath: serializedSuiteMetadataOutputFilepath,
		shutdownChan: shutdownChan,
	}
}

func (service *suiteMetadataSerializationService) HandleSuiteRegistrationEvent() error {
	go func() {
		time.Sleep(serializationTimeout)
		service.mutex.Lock()
		defer service.mutex.Unlock()
		if !service.hasSerializeBeenCalled {
			service.shutdownChan <- api_container_exit_codes.NoTestExecutionRegistered
		}
	}()
	return nil
}

func (service *suiteMetadataSerializationService) HandlePostShutdownEvent() error {
	// No cleanup needed
	return nil
}

func (service *suiteMetadataSerializationService) SerializeSuiteMetadata(
		ctx context.Context,
		apiSuiteMetadata *bindings.TestSuiteMetadata) (*emptypb.Empty, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	if service.hasSerializeBeenCalled {
		// We don't use stacktrace to not leak internal details about the API container to callers
		return nil, errors.New("suite metadata serialization has already been called; it should not be called twice")
	}
	service.hasSerializeBeenCalled = true

	initializerAcceptableSuiteMetadata := convertToInitializerMetadata(apiSuiteMetadata)

	logrus.Debugf(
		"Printing test suite metadata to file '%v'...",
		service.serializedSuiteMetadataOutputFilepath)
	if err := printSuiteMetadataToFile(
			initializerAcceptableSuiteMetadata,
			service.serializedSuiteMetadataOutputFilepath); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the suite metadata object to file '%v'", service.serializedSuiteMetadataOutputFilepath)
	}
	logrus.Debugf("Successfully serialized suite metadata to file")

	defer func() {
		service.shutdownChan <- api_container_exit_codes.SuccessfulExit
	}()

	return &emptypb.Empty{}, nil
}

func convertToInitializerMetadata(apiSuiteMetadata *bindings.TestSuiteMetadata) test_suite_metadata_acquirer.TestSuiteMetadata {
	allInitializerAcceptableTestMetadata := map[string]test_suite_metadata_acquirer.TestMetadata{}
	for testName, apiTestMetadata := range apiSuiteMetadata.TestMetadata {
		initializerAcceptableTestMetadata := test_suite_metadata_acquirer.NewTestMetadata(
			apiTestMetadata.IsPartitioningEnabled,
			apiTestMetadata.UsedArtifactUrls)

		allInitializerAcceptableTestMetadata[testName] = *initializerAcceptableTestMetadata
	}

	initializerAcceptableSuiteMetadata := test_suite_metadata_acquirer.NewTestSuiteMetadata(
		apiSuiteMetadata.NetworkWidthBits,
		allInitializerAcceptableTestMetadata)
	return *initializerAcceptableSuiteMetadata
}

func printSuiteMetadataToFile(suiteMetadata test_suite_metadata_acquirer.TestSuiteMetadata, filepath string) error {
	fp, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening the file '%v' to write suite metadata JSON to", filepath)
	}
	defer fp.Close()

	bytes, err := json.Marshal(suiteMetadata)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing suite metadata object to JSON")
	}

	if _, err := fp.Write(bytes); err != nil {
		return stacktrace.Propagate(err, "An error occurred writing the suite metadata JSON string to file '%v'", filepath)
	}

	return nil
}

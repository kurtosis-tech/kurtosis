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
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"os"
	"sync"
)

type suiteMetadataSerializationService struct {
	mutex                                 *sync.Mutex
	hasSerializeBeenCalled                bool
	serializedSuiteMetadataOutputFilepath string
	shutdownChan chan exit_codes.ApiContainerExitCode
}

func newSuiteMetadataSerializationService(
		serializedSuiteMetadataOutputFilepath string,
		shutdownChan chan exit_codes.ApiContainerExitCode) *suiteMetadataSerializationService {
	return &suiteMetadataSerializationService{
		mutex: &sync.Mutex{},
		hasSerializeBeenCalled: false,
		serializedSuiteMetadataOutputFilepath: serializedSuiteMetadataOutputFilepath,
		shutdownChan: shutdownChan,
	}
}

func (service *suiteMetadataSerializationService) HandleSuiteRegistrationEvent() error {
	// Nothing to do here
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

	return &emptypb.Empty{}, nil
}

func convertToInitializerMetadata(apiSuiteMetadata *bindings.TestSuiteMetadata) test_suite_metadata_acquirer.TestSuiteMetadata {
	allInitializerAcceptableTestMetadata := map[string]test_suite_metadata_acquirer.TestMetadata{}
	for testName, apiTestMetadata := range apiSuiteMetadata.TestMetadata {
		artifactIdToUrl := map[string]string{}
		for artifactUrl := range apiTestMetadata.UsedArtifactUrls {
			artifactId := generateArtifactId(artifactUrl)
			artifactIdToUrl[artifactId] = artifactUrl
		}

		initializerAcceptableTestMetadata := test_suite_metadata_acquirer.NewTestMetadata(
			apiTestMetadata.IsPartitioningEnabled,
			// TODO reconsider whether we even want artifact IDs at all
			artifactIdToUrl)

		allInitializerAcceptableTestMetadata[testName] = *initializerAcceptableTestMetadata
	}

	initializerAcceptableSuiteMetadata := test_suite_metadata_acquirer.NewTestSuiteMetadata(
		apiSuiteMetadata.NetworkWidthBits,
		allInitializerAcceptableTestMetadata)
	return *initializerAcceptableSuiteMetadata
}

// TODO Write tests for this by splitting it into metadata-generating function and writing function
//  then testing the metadata-generating
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

// TODO replace with proper aritfact ID-generating function from kurt-go
func generateArtifactId(artifactUrl string) string {
	return artifactUrl
}
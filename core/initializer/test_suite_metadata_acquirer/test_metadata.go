/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"github.com/palantir/stacktrace"
	"strings"
)

type TestMetadata struct {
	IsPartitioningEnabled bool	`json:"isPartitioningEnabled"`

	// A "set" of all the artifact URLs that the test wants, which the initializer will
	//  download and make ready for the test at runtime
	UsedArtifacts map[string]bool `json:"usedArtifacts"`

	TestSetupTimeout uint32
	TestExecutionTimeout uint32
}


// Even though the struct's fields must be public for JSON, we create a constructor so that we don't forget
//  to initialize any fields
func NewTestMetadata(isPartitioningEnabled bool, usedArtifacts map[string]bool) *TestMetadata {
	return &TestMetadata{IsPartitioningEnabled: isPartitioningEnabled, UsedArtifacts: usedArtifacts}
}

// Go stupidly doesn't have any way to require JSON fields, so we have to manually do it
func validateTestMetadata(testMetadata TestMetadata) error {
	for artifactUrl := range testMetadata.UsedArtifacts {
		if len(strings.TrimSpace(artifactUrl)) == 0 {
			return stacktrace.NewError("Found empty used artifact URL: %v", artifactUrl)
		}
	}
	return nil
}

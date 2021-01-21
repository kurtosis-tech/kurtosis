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

	// A map of all the artifacts that the test wants,
	//  which the initializer will download and make ready for the test at runtime
	// The map is in the form of ID -> URL, where the ID is:
	//	1. The ID that the initializer should associate the artifact with after downloading it and
	//	2. The ID that the client will use to retrieve the artifact when a test requests it
	UsedArtifacts map[string]string `json:"usedArtifacts"`
}

// Even though the struct's fields must be public for JSON, we create a constructor so that we don't forget
//  to initialize any fields
func NewTestMetadata(isPartitioningEnabled bool, usedArtifacts map[string]string) *TestMetadata {
	return &TestMetadata{IsPartitioningEnabled: isPartitioningEnabled, UsedArtifacts: usedArtifacts}
}

// Go stupidly doesn't have any way to require JSON fields, so we have to manually do it
func validateTestMetadata(testMetadata TestMetadata) error {
	for artifactId, artifactUrl := range testMetadata.UsedArtifacts {
		if len(strings.TrimSpace(artifactId)) == 0 {
			return stacktrace.NewError("Found empty used artifact ID: %v", artifactId)
		}
		if len(strings.TrimSpace(artifactUrl)) == 0 {
			return stacktrace.NewError("Found empty used artifact URL: %v", artifactUrl)
		}
	}
	return nil
}

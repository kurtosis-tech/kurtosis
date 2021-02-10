/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

// Simple package struct to contain information about a test suite
type TestSuiteMetadata struct {
	// How many bits to give each test subnetwork
	NetworkWidthBits uint32		`json:"networkWidthBits"`

	TestMetadata map[string]TestMetadata `json:"testMetadata"`
}

// Even though the struct's fields must be public for JSON, we create a constructor so that we don't forget
//  to initialize any fields
func NewTestSuiteMetadata(networkWidthBits uint32, testMetadata map[string]TestMetadata) *TestSuiteMetadata {
	return &TestSuiteMetadata{NetworkWidthBits: networkWidthBits, TestMetadata: testMetadata}
}

// TODO switch to member function?
// Go stupidly doesn't have any way to require JSON fields, so we have to manually do it
func validateTestSuiteMetadata(suiteMetadata TestSuiteMetadata) error {
	if suiteMetadata.NetworkWidthBits == 0 {
		return stacktrace.NewError("Test suite metdata has a network width bits == 0")
	}
	if suiteMetadata.TestMetadata == nil {
		return stacktrace.NewError("Test metadata map is nil")
	}
	if len(suiteMetadata.TestMetadata) == 0 {
		return stacktrace.NewError("Test suite doesn't declare any tests")
	}
	for testName, testMetadata := range suiteMetadata.TestMetadata {
		if len(strings.TrimSpace(testName)) == 0 {
			return stacktrace.NewError("Test name '%v' is empty", testName)
		}
		if err := validateTestMetadata(testMetadata); err != nil {
			return stacktrace.Propagate(err, "An error occurred validating metadata for test '%v'", testName)
		}
		logrus.Infof("Test metadata: %+v", testMetadata)
	}
	return nil
}

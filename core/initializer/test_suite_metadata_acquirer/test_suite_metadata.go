/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_metadata_acquirer

// Simple package struct to contain information about a test suite
type TestSuiteMetadata struct {
	// How many bits to give each test subnetwork
	NetworkWidthBits uint32

	TestNames map[string]bool
}


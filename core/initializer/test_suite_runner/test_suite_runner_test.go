/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis/test_suite/rpc_api/bindings"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestBlockedExecutionWhenNoPerms(t *testing.T) {
	suiteMetadata := getTestingSuiteMetadata(1)
	perms := permissions.FromPermissionsSet(map[string]bool{})
	result, err := RunTests(
		perms,
		"1234-abcd",
		nil,
		nil,
		suiteMetadata,
		map[string]bool{},
		1,
		nil,
		nil)
	assert.False(t, result)
	assert.Contains(t, err.Error(), suiteExecutionPermissionDeniedErrStr)
}

func TestBlockedExecutionWhenRestrictedPerms(t *testing.T) {
	suiteMetadata := getTestingSuiteMetadata(4)
	perms := permissions.FromPermissionsSet(map[string]bool{
		permissions.RestrictedTestExecutionPermission: true,
	})
	result, err := RunTests(
		perms,
		"1234-abcd",
		nil,
		nil,
		suiteMetadata,
		map[string]bool{},
		1,
		nil,
		nil)
	assert.False(t, result)
	assert.Contains(t, err.Error(), suiteExecutionPermissionDeniedErrStr)
}

func getTestingSuiteMetadata(numTests int) *bindings.TestSuiteMetadata {
	testMetadata := map[string]*bindings.TestMetadata{}
	for i := 0; i < numTests; i++ {
		testMetadata["test" + strconv.Itoa(i)] = &bindings.TestMetadata{
			IsPartitioningEnabled: 		false,
			UsedArtifactUrls:        	map[string]bool{},
		}
	}
	return &bindings.TestSuiteMetadata{
		TestMetadata: testMetadata,
		NetworkWidthBits: 0,
	}
}
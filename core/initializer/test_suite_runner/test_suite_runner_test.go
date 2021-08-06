/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/kurtosis_testsuite_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
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
		"5678-efgh",
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
		"5678-efgh",
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

func getTestingSuiteMetadata(numTests int) *kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata {
	testMetadata := map[string]*kurtosis_testsuite_rpc_api_bindings.TestMetadata{}
	for i := 0; i < numTests; i++ {
		testMetadata["test" + strconv.Itoa(i)] = &kurtosis_testsuite_rpc_api_bindings.TestMetadata{
			IsPartitioningEnabled:     false,
			TestSetupTimeoutInSeconds: 60,
			TestRunTimeoutInSeconds:   60,
		}
	}
	return &kurtosis_testsuite_rpc_api_bindings.TestSuiteMetadata{
		TestMetadata: testMetadata,
	}
}

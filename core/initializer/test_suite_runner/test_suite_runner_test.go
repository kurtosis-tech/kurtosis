/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_suite_runner

import (
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis/initializer/test_suite_metadata_acquirer"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestBlockedExecutionWhenNoPerms(t *testing.T) {
	suiteMetadata := getTestingSuiteMetadata(1)
	perms := permissions.FromPermissionsSet(map[string]bool{})
	result, err := RunTests(
		perms,
		uuid.New(),
		nil,
		nil,
		suiteMetadata,
		map[string]bool{},
		1,
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
		uuid.New(),
		nil,
		nil,
		suiteMetadata,
		map[string]bool{},
		1,
		nil)
	assert.False(t, result)
	assert.Contains(t, err.Error(), suiteExecutionPermissionDeniedErrStr)
}

func getTestingSuiteMetadata(numTests int) test_suite_metadata_acquirer.TestSuiteMetadata {
	testMetadata := map[string]test_suite_metadata_acquirer.TestMetadata{}
	for i := 0; i < numTests; i++ {
		testMetadata["test" + strconv.Itoa(i)] = test_suite_metadata_acquirer.TestMetadata{
		IsPartitioningEnabled: false,
			UsedArtifacts:         nil,
		}
	}
	return test_suite_metadata_acquirer.TestSuiteMetadata{
		NetworkWidthBits: 0,
		TestMetadata:     testMetadata,
	}
}
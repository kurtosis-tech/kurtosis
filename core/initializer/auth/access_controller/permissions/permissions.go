/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package permissions

import (
	"github.com/kurtosis-tech/kurtosis/commons/user_support_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"math"
)

const (
	// The maximum number of tests allowed in a test suite when the restricted test execution perm is the only
	//  one the user has
	maxTestsAllowedWhenRestricted = 3

	maxTestsAllowedWhenUnlimited = math.MaxInt32

	// Test execution permissions
	RestrictedTestExecutionPermission = "execute-tests:restricted"
	UnlimitedTestExecutionPermission  = "execute-tests:unlimited"
)

type Permissions struct {
	maxNumTestsAllowed int
}

// Static factory method
func FromPermissionsSet(permissionsSet map[string]bool) *Permissions {
	// By default, no tests are allowed
	maxNumTestsAllowed := 0
	for permissionStr := range permissionsSet {
		switch permissionStr {
		case RestrictedTestExecutionPermission:
			maxNumTestsAllowed = maxInt(maxNumTestsAllowed, maxTestsAllowedWhenRestricted)
		case UnlimitedTestExecutionPermission:
			maxNumTestsAllowed = maxInt(maxNumTestsAllowed, maxTestsAllowedWhenUnlimited)
		default:
			// We don't throw an error here to allow us to add permissions in the future, transparent to users
			logrus.Debugf("Unrecognized permission: %v", permissionStr)
			continue
		}
	}

	return &Permissions{
		maxNumTestsAllowed: maxNumTestsAllowed,
	}
}

func (perms *Permissions) CanExecuteSuite(numTestsInSuite int) error {
	if numTestsInSuite > perms.maxNumTestsAllowed {
		return stacktrace.NewError("Your current Kurtosis license only allows for testsuites with %v tests " +
			"and you're trying to run a testsuite with %v tests; either upgrade your Kurtosis license by emailing " +
			"%v or reduce the number of tests in your testsuite",
			perms.maxNumTestsAllowed,
			numTestsInSuite,
			user_support_constants.InfoEmail)
	}
	return nil
}

// Why, Go... why does your stdlib suck
func maxInt(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}


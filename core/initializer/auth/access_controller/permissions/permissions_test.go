/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package permissions

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaxInt(t *testing.T) {
	max := maxInt(32, 16)
	assert.Equal(t, 32, max)

	max2 := maxInt(12, 29)
	assert.Equal(t, 29, max2)
}

func TestNoTestsAllowed(t *testing.T) {
	noPerms := FromPermissionsSet(map[string]bool{})
	assert.Equal(t, 0, noPerms.maxNumTestsAllowed)

	restricted := FromPermissionsSet(map[string]bool{RestrictedTestExecutionPermission: true})
	assert.Equal(t, maxTestsAllowedWhenRestricted, restricted.maxNumTestsAllowed)

	unlimited := FromPermissionsSet(map[string]bool{UnlimitedTestExecutionPermission: true})
	assert.Equal(t, maxTestsAllowedWhenUnlimited, unlimited.maxNumTestsAllowed)

	// Verify that multiple perms allows the highest
	unlimitedWithRestricted := FromPermissionsSet(map[string]bool{
		RestrictedTestExecutionPermission: true,
		UnlimitedTestExecutionPermission:  true,
	})
	assert.Equal(t, maxTestsAllowedWhenUnlimited, unlimitedWithRestricted.maxNumTestsAllowed)
}

func TestCanExecuteSuite(t *testing.T) {
	limit := 5
	perms := Permissions{maxNumTestsAllowed: limit}
	assert.Nil(t, perms.CanExecuteSuite(0))
	assert.Nil(t, perms.CanExecuteSuite(limit - 1))
	assert.Nil(t, perms.CanExecuteSuite(limit))
	assert.NotNil(t, perms.CanExecuteSuite(limit + 1))
	assert.NotNil(t, perms.CanExecuteSuite(limit + 10))
}

func TestNoIssueOnUnrecognizedPerm(t *testing.T) {
	// We don't want to error on an unrecognized perm to allow us to transparently roll out new perms
	// Verify that multiple perms allows the highest
	FromPermissionsSet(map[string]bool{
		"this-permission-doesnt-exist": true,
	})
}


/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package access_controller

import (
	permissions3 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/access_controller/permissions"
	auth0_constants2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/auth0_constants"
	test_mocks2 "github.com/kurtosis-tech/kurtosis/cli/commands/test/testing_machinery/auth/test_mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTokenClaimsToPermissions(t *testing.T) {
	audience := auth0_constants2.Audience
	issuer := auth0_constants2.Issuer
	expiredInSeconds := 3600
	permissions := []string{
		permissions3.RestrictedTestExecutionPermission,
		permissions3.UnlimitedTestExecutionPermission,
	}

	token, err := test_mocks2.CreateTestToken(
		test_mocks2.TestAuth0PrivateKey,
		audience,
		issuer,
		expiredInSeconds,
		permissions,
	)
	assert.Nil(t, err)

	parsedClaims, err := parseTokenClaims(
		map[string]string{
			test_mocks2.TestAuth0KeyId: test_mocks2.TestAuth0PublicKeys[test_mocks2.TestAuth0KeyId],
		},
		token)
	assert.Nil(t, err)

	assert.Equal(t, audience, parsedClaims.Audience)
	assert.Equal(t, issuer, parsedClaims.Issuer)
	assert.Equal(t, permissions, parsedClaims.Permissions)

	parsePermissionsFromClaims(parsedClaims)
}

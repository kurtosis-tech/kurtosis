/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	permissions2 "github.com/kurtosis-tech/kurtosis/initializer/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/test_mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTokenClaimsToPermissions(t *testing.T) {
	audience := auth0_constants.Audience
	issuer := auth0_constants.Issuer
	expiredInSeconds := 3600
	permissions := []string{
		permissions2.RestrictedTestExecutionPermission,
		permissions2.UnlimitedTestExecutionPermission,
	}

	token, err := test_mocks.CreateTestToken(
		test_mocks.TestAuth0PrivateKey,
		audience,
		issuer,
		expiredInSeconds,
		permissions,
	)
	assert.Nil(t, err)

	parsedClaims, err := parseTokenClaims(
		map[string]string{
			test_mocks.TestAuth0KeyId: test_mocks.TestAuth0PublicKeys[test_mocks.TestAuth0KeyId],
		},
		token)
	assert.Nil(t, err)

	assert.Equal(t, audience, parsedClaims.Audience)
	assert.Equal(t, issuer, parsedClaims.Issuer)
	assert.Equal(t, permissions, parsedClaims.Permissions)

	parsePermissionsFromClaims(parsedClaims)
}

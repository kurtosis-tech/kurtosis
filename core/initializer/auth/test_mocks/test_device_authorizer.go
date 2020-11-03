/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_mocks

import "github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authorizers"

type TestDeviceAuthorizer struct {
	ThrowErrorOnAuthorize bool
	ScopeToReturn string
	ExpiresInSeconds int
}

func (t TestDeviceAuthorizer) AuthorizeUserDevice() (*auth0_authorizers.TokenResponse, error) {
	panic("implement me")
}


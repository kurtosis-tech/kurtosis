/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_mocks

import "github.com/kurtosis-tech/stacktrace"

type MockClientCredentialsAuthorizer struct{
	throwErrorOnAuthorize bool
	tokenToReturn string
}

func NewMockClientCredentialsAuthorizer(
		throwErrorOnAuthorize bool,
		tokenToReturn string) *MockClientCredentialsAuthorizer {
	return &MockClientCredentialsAuthorizer{
		throwErrorOnAuthorize: throwErrorOnAuthorize,
		tokenToReturn: tokenToReturn,
	}
}

func (m MockClientCredentialsAuthorizer) AuthenticateClientCredentials(clientId string, clientSecret string) (string, error) {
	if m.throwErrorOnAuthorize {
		return "", stacktrace.NewError("Test error on authorization, as requested")
	}
	return m.tokenToReturn, nil
}



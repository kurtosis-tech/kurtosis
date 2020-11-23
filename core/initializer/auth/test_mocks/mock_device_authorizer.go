/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_mocks

import (
	"github.com/palantir/stacktrace"
)

type MockDeviceAuthorizer struct {
	throwErrorOnAuthorize bool
	tokenToReturn string
	scopeToReturn string
	expiresInSeconds int
}

func NewMockDeviceAuthorizer(
		throwErrorOnAuthorize bool,
		tokenToReturn string) *MockDeviceAuthorizer {
	return &MockDeviceAuthorizer{
		throwErrorOnAuthorize: throwErrorOnAuthorize,
		tokenToReturn: tokenToReturn,
	}
}



func (t MockDeviceAuthorizer) AuthorizeUserDevice() (string, error) {
	if t.throwErrorOnAuthorize {
		return "", stacktrace.NewError("Test error on authorization, as requested")
	}
	return t.tokenToReturn, nil
}


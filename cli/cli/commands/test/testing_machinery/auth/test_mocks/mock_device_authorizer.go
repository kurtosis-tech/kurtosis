/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package test_mocks

import (
	"github.com/kurtosis-tech/stacktrace"
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



func (t MockDeviceAuthorizer) AuthorizeDeviceAndAuthenticate() (string, error) {
	if t.throwErrorOnAuthorize {
		return "", stacktrace.NewError("Test error on authorization, as requested")
	}
	return t.tokenToReturn, nil
}


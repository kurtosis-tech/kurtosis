/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_mocks

import "github.com/palantir/stacktrace"

const (
	testClientCredAuthorizerPrivateKey = "TODO TODO TODO"
	TestClientCredAuthorizerKeyId = "test-key-id"
	TestClientCredAuthorizerPubKey = "TODO TODO TODO"
)

type TestClientCredentialsAuthorizer struct{
	ThrowErrorOnAuthorize bool
	ScopeToReturn string
	ExpiresInSeconds int
}

func (t TestClientCredentialsAuthorizer) AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error) {
	if t.ThrowErrorOnAuthorize {
		return nil, stacktrace.NewError("TEST ERROR")
	}

	// TODO create a token with the private key!
	tokenStr := "TODO TODO THIS IS BROKEN"

	TokenResponse{
		AccessToken: tokenStr,
		Scope:       "",
		ExpiresIn:   0,
		TokenType:   "",
	}
	return
}

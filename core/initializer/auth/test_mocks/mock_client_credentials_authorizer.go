/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_mocks

const (
	testClientCredAuthorizerPrivateKey = "TODO TODO TODO"
	TestClientCredAuthorizerKeyId = "test-key-id"
	TestClientCredAuthorizerPubKey = "TODO TODO TODO"
)

type MockClientCredentialsAuthorizer struct{
	throwErrorOnAuthorize bool
	scopeToReturn string
	expiresInSeconds int
}

/*
func (t MockClientCredentialsAuthorizer) AuthorizeClientCredentials(clientId string, clientSecret string) (*TokenResponse, error) {
	if t.throwErrorOnAuthorize {
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


 */
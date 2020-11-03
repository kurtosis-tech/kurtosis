/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_mocks

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_token_claims"
)

const (
	TestAuth0KeyId = "test-key-id"
	TestAuth0PrivateKey = "TODO TODO TODO"
)

var TestAuth0PublicKeys = map[string]string{
	TestAuth0KeyId: "TODO TODO TODO",
}

func CreateTestToken() {
	Auth0
	GetValid
	auth0_token_claims.Auth0TokenClaims{
		Audience:  auth0_constants.Audience,
		ExpiresAt: 0,
		IssuedAt:  0,
		Issuer:    auth0_constants.Issuer,
		Scope:     "",
		Subject:   "",
	}
	jwt.NewWithClaims()
}
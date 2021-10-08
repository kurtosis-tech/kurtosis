/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package auth0_token_claims

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/auth0_constants"
	"github.com/palantir/stacktrace"
)

type Auth0TokenClaims struct {
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	Scope 	  string `json:"scope"`
	Subject   string `json:"sub,omitempty"`
	Permissions []string	`json:"permissions"`
}

func (claims Auth0TokenClaims) Valid() error {
	// We intentionally don't check the token expiration here because if a token were expired, we could only throw an error
	// An error here means the entire token is invalid and should be rejected, but we want to retry if a token is expired
	// Instead, we check the expiration after the token is parsed

	if claims.Audience != auth0_constants.Audience {
		return stacktrace.NewError("Claims audience '%v' != expected audience '%v'", claims.Audience, auth0_constants.Audience)
	}

	if claims.Issuer != auth0_constants.Issuer {
		return stacktrace.NewError("Claims issuer '%v' != expected issuer '%v'", claims.Issuer, auth0_constants.Issuer)
	}

	return nil
}

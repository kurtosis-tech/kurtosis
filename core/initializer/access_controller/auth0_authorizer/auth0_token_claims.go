/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package auth0_authorizer

import (
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_constants"
	"github.com/palantir/stacktrace"
	"time"
)

const (
	// Users maybe be out of internet range when their token expires, so we give them a grace period to
	//  get a new token so they don't get surprised by it
	tokenExpirationGracePeriod = 5 * 24 * time.Hour
)

type Auth0TokenClaims struct {
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	Scope 	  string `json:"scope"`
	Subject   string `json:"sub,omitempty"`
}

func (claims Auth0TokenClaims) Valid() error {
	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt, 0)

	// We give users a grace period because they may not have internet connection when their token expires
	if expiration.Add(tokenExpirationGracePeriod).Before(now) {
		return stacktrace.NewError(
			"Token claim expires at '%v', which is more than the grace period (%v) ago",
			expiration,
			tokenExpirationGracePeriod)
	}

	if claims.Audience != auth0_constants.Audience {
		return stacktrace.NewError("Claims audience '%v' != expected audience '%v'", claims.Audience, auth0_constants.Audience)
	}

	if claims.Issuer != auth0_constants.Issuer {
		return stacktrace.NewError("Claims issuer '%v' != expected issuer '%v'", claims.Issuer, auth0_constants.Issuer)
	}

	return nil
}

func (claims Auth0TokenClaims) GetGracePeriod() time.Duration {
	return tokenExpirationGracePeriod
}
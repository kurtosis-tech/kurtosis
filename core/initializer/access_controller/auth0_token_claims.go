/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_constants"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
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

	logrus.Debugf("Expiration: %v", expiration)
	logrus.Debugf("Grace period: %v", tokenExpirationGracePeriod)
	logrus.Debugf("Expiration + grace period: %v", expiration.Add(tokenExpirationGracePeriod))
	logrus.Debugf("Expiration + grace period before now: %v", expiration.Add(tokenExpirationGracePeriod).Before(now))

	// We give users a grace period because they may not have internet connection when their token expires
	if expiration.Add(tokenExpirationGracePeriod).Before(now) {
		return stacktrace.NewError(
			"Token claim expires at '%v', which is beyond the grace period of %v",
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


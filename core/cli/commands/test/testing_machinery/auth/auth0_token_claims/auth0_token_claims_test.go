/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package auth0_token_claims

import (
	"github.com/kurtosis-tech/kurtosis-core/cli/commands/test/testing_machinery/auth/auth0_constants"
	"testing"
	"time"
)

func TestValidToken(t *testing.T) {
	claims := getValidClaim()
	if err := claims.Valid(); err != nil {
		t.Fatal("Claim validation failed when it should have succeeded")
	}
}

// The expiration date should NOT be done inside the claims.Valid() method, because then we can't implement
//  a grace period
func TestExpirationDateNotImplemented(t *testing.T) {
	claims := getValidClaim()
	claims.ExpiresAt = 0
	if err := claims.Valid(); err != nil {
		t.Fatal("Token claim validation should NOT fail validation if it's past expiration")
	}
}

func TestInvalidIssuer(t *testing.T) {
	claims := getValidClaim()
	claims.Issuer = "Dwayne Johnson"
	if err := claims.Valid(); err == nil {
		t.Fatal("Token did not fail validation on invalid issuer")
	}
}

func TestInvalidAudience(t *testing.T) {
	claims := getValidClaim()
	claims.Audience = "Your Mom"
	if err := claims.Valid(); err == nil {
		t.Fatal("Token did not fail validation on invalid audience")
	}
}

func getValidClaim() Auth0TokenClaims {
	now := time.Now()
	oneHourFromNow := now.Add(24 * time.Hour)
	oneHourFromNow.Unix()
	return Auth0TokenClaims{
		Audience:    auth0_constants.Audience,
		ExpiresAt:   oneHourFromNow.Unix(),
		IssuedAt:    now.Unix(),
		Issuer:      auth0_constants.Issuer,
		Scope:       "",
		Subject:     "not-sure-what-goes-here",
		Permissions: []string{},
	}
}

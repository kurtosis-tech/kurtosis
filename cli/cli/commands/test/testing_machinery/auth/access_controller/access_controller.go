/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/auth0_token_claims"
	"github.com/kurtosis-tech/stacktrace"
	"time"
)

const (
	// Users maybe be out of internet range when their token expires, so we give them a grace period to
	//  get a new token so they don't get surprised by it
	tokenExpirationGracePeriod = 5 * 24 * time.Hour

	// How long we'll pause after displaying auth warnings, to give users a chance to see it
	authWarningPause = 3 * time.Second

	// For extra security, make sure only the user can read & write the session cache file
	sessionCacheFilePerms = 0600

	// Key in the Headers hashmap of the token that points to the key ID
	keyIdTokenHeaderKey = "kid"
)

// NOTE: If we wanted, we could also refactor this into "abstract class" syntax, where it contains a couple
//  a copule interfaces that are responsible for each little bit of functionality
type AccessController interface {
	Authenticate() (*permissions.Permissions, error)
}



// ============================== HELPER FUNCTIONS =========================================

// jwt-go requires that you have a function which uses the token to get the public key for validating the tokne
// This is our method for generating that function for use with Kurtosis JWT tokens
func generatePubKeyExtractorFunc(rsaPubKeysPem map[string]string) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		// IMPORTANT: Validating the algorithm per https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, stacktrace.NewError(
				"Expected token algorithm '%v' but got '%v'",
				jwt.SigningMethodRS256.Name,
				token.Header)
		}

		untypedKeyId, found := token.Header[keyIdTokenHeaderKey]
		if !found {
			return nil, stacktrace.NewError("No key ID key '%v' found in token header", keyIdTokenHeaderKey)
		}
		keyId, ok := untypedKeyId.(string)
		if !ok {
			return nil, stacktrace.NewError("Found key ID, but value was not a string")
		}

		pubKeyPemStr, found := rsaPubKeysPem[keyId]
		if !found {
			return nil, stacktrace.NewError("No public RSA key found corresponding to key ID from token '%v'", keyId)
		}

		pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pubKeyPemStr))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred parsing the public key base64 for key ID '%v'; this is a code bug", keyId)
		}

		return pubKey, nil
	}
}

/*
Parses a token string, validates the claims, and returns the claims object

Args:
	rsaPubKeysPem: The token will be considered valid if it was signed by an RSA private key corresponding to one of
		these public keys, with the key of this map being the "key ID", which can be found in the token's "kid" header
	tokenStr: The signed token string
 */

func parseTokenClaims(rsaPubKeysPem map[string]string, tokenStr string) (*auth0_token_claims.Auth0TokenClaims, error) {
	token, err := new(jwt.Parser).ParseWithClaims(
		tokenStr,
		&auth0_token_claims.Auth0TokenClaims{},
		generatePubKeyExtractorFunc(rsaPubKeysPem),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing or validating the JWT token")
	}

	claims, ok := token.Claims.(*auth0_token_claims.Auth0TokenClaims)
	if !ok {
		return nil, stacktrace.NewError("Could not cast token claims to Auth0 token claims object, indicating an invalid token")
	}

	return claims, nil
}

func parsePermissionsFromClaims(claims *auth0_token_claims.Auth0TokenClaims) (*permissions.Permissions) {
	permsSet := map[string]bool{}
	for _, permStr := range claims.Permissions {
		permsSet[permStr] = true
	}
	result := permissions.FromPermissionsSet(permsSet)
	return result
}

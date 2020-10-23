/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_authorizer"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/encrypted_session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// Users maybe be out of internet range when their token expires, so we give them a grace period to
	//  get a new token so they don't get surprised by it
	tokenExpirationGracePeriod = 0 * 24 * time.Hour

	// How long we'll pause after displaying auth warnings, to give users a chance to see it
	authWarningPause = 3 * time.Second

	// For extra security, make sure only the user can read & write the session cache file
	sessionCacheFilePerms = 0600

	// We need to verify that the token has the expected algorithm, else we have a security vulnerability:
	//  https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	expectedTokenHeaderAlgorithm = "RS256"

	// Key in the Headers hashmap of the token that points to the key ID
	keyIdTokenHeaderKey = "kid"
)

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them for their username and password

Args:
	sessionCacheFilepath: Filepath to store the encrypted session cache at

Returns:
	An error if and only if an irrecoverable login error occurred
 */
func RunDeveloperMachineAuthFlow(sessionCacheFilepath string) error {
	cache := encrypted_session_cache.NewEncryptedSessionCache(sessionCacheFilepath, sessionCacheFilePerms)

	tokenStr, err := getTokenStr(cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	claims, err := parseAndValidateTokenClaims(tokenStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	scope, err := getScopeFromClaimsAndRenewIfNeeded(claims, cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred renewing the token or getting the scope from the token's claims")
	}

	if scope != auth0_constants.ExecutionScope {
		return stacktrace.NewError(
			"Kurtosis requires scope '%v' to run but token has scope '%v'; this is most likely due to an expired Kurtosis license",
			auth0_constants.ExecutionScope,
			scope)
	}
	return nil
}

/*
This workflow is for authenticating and authorizing Kurtosis tests running in CI (no device or username).
	See also: https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
 */
func RunCIAuthFlow(clientId string, clientSecret string) error {
	tokenResponse, err := auth0_authorizer.AuthorizeClientCredentials(clientId, clientSecret)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred authenticating with the client ID & secret")
	}

	claims, err := parseAndValidateTokenClaims(tokenResponse.AccessToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing and validating the token claims")
	}

	scope := claims.Scope
	if scope != auth0_constants.ExecutionScope {
		return stacktrace.NewError(
			"Kurtosis requires scope '%v' to run but token has scope '%v'; this is most likely due to an expired Kurtosis license",
			auth0_constants.ExecutionScope,
			scope)
	}
	return nil
}

// ============================== PRIVATE HELPER FUNCTIONS =========================================
/*
Gets the token string, either by reading a valid cache or by prompting the user for their login credentials
 */
func getTokenStr(cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	var result string
	session, err := cache.LoadSession()
	if err != nil {
		// We couldn't load any cached session, so the user MUST log in
		logrus.Tracef("The following error occurred loading the session from file: %v", err)
		tokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred during Auth0 authentication")
		}

		// The user has successfully authenticated, so we're good to go
		newSession := encrypted_session_cache.Session{
			Token:                    tokenResponse.AccessToken,
		}
		if err := cache.SaveSession(newSession); err != nil {
			logrus.Warnf("We received a token from Auth0 but the following error occurred when caching it locally:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Warn("If this error isn't corrected, you'll need to log into Kurtosis every time you run it")
			time.Sleep(authWarningPause)
		}
		result = tokenResponse.AccessToken
	} else {
		// We were able to load a session
		result = session.Token
	}

	return result, nil
}

func parseAndValidateTokenClaims(tokenStr string) (Auth0TokenClaims, error) {
	// This includes validation like expiration date, issuer, etc.
	// See Auth0TokenClaims for more details

	// TODO POTENTIAL SECURITY HOLE: we don't validate token signatures!!!!! The reason this isn't *so* catastrophic:
	//  1) to support offline Kurtosis operation, we need to be able to validate tokens offline
	//  2) unfortunately, validating tokens requires online access (to pull the public keys for verifying the token)
	//  3) we'd have to cache the public keys locally, but we'd need to encrypt them to make sure the user can't mess with them
	//  4) writing the cert-pulling-and-encrypting logic is a big hassle, so instead we just encrypt the token (for the same effect)
	//  The action item is to find a secure public key store that can't be tampered with, and to pull the public certs
	token, _, err := new(jwt.Parser).ParseUnverified(
		tokenStr,
		&Auth0TokenClaims{},
		// This is the "key extractor" algorithm. It should use the "kid" (key ID) field to determine the
		//  private key that the token was signed with (and therefore which public key to use)
		// Only uncomment this when we switch back to validating tokens
		/*
		func(token *jwt.Token) (interface{}, error) {
			// IMPORTANT: Validating the algorithm per https://godoc.org/github.com/dgrijalva/jwt-go#example-Parse--Hmac
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, stacktrace.NewError(
					"Expected token algorithm '%v' but got '%v'",
					expectedTokenHeaderAlgorithm,
					token.Header)
			}

			// TOOD pull the Auth0 public keys
			// TOOD return the one corresponding to the "kid" header value
		},
		 */
	)

	if err != nil {
		return Auth0TokenClaims{}, stacktrace.Propagate(err, "An error occurred parsing or validating the JWT token")
	}

	// NOTE: At this point we SHOULD be checking the signature by comparing against the Auth0 public
	//  certs (see https://auth0.com/docs/quickstart/backend/golang ). However, we want to let users be able to
	//  run Kurtosis while offline, which means we'd need to cache the Auth0 public keys. If we store the
	//  public keys locally, we'd need to encrypt them. If we're decrypting them locally, the encryption key has
	//  to be stored in code. Therefore, encrypting the token itself and not doing any signature checks gives
	//  the same result for less complexity.
	// To actually do the validation, use ParseWithClaims and check token.Valid:
	//	https://godoc.org/github.com/dgrijalva/jwt-go#example-ParseWithClaims--CustomClaimsType

	claims, ok := token.Claims.(*Auth0TokenClaims)
	if !ok {
		return Auth0TokenClaims{}, stacktrace.NewError("Could not cast token claims to Auth0 token claims object, indicating an invalid token")
	}

	return *claims, nil
}

func getScopeFromClaimsAndRenewIfNeeded(claims Auth0TokenClaims, cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt, 0)
	if expiration.After(now) {
		return claims.Scope, nil
	}

	// If we've gotten here, it means that the token is beyond the expiration but not beyond the grace period (else
	//  token validation would have failed completely)
	expirationExceededAmount := now.Sub(expiration)
	logrus.Infof("Kurtosis token expired %v ago; attempting to get a new token...", expirationExceededAmount)
	newTokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
	if err != nil {
		logrus.Debugf("Token expiration error: %v", err)
		logrus.Warnf(
			"WARNING: Your Kurtosis token expired %v ago and we couldn't reach Auth0 to get a new one",
			expirationExceededAmount)
		logrus.Warnf(
			"You will have a grace period of %v from expiration to get a connection to Auth0 before Kurtosis stops working",
			tokenExpirationGracePeriod)
		time.Sleep(authWarningPause)

		// NOTE: If it's annoying for users for Kurtosis to try and hit Auth0 on every run after their token is expired
		//  (say, they have to wait for the connection to time out) then we can add a tracker in the session on the last
		//  time we warned them and only warn them say every 3 hours

		return claims.Scope, nil
	}

	// If we've gotten here, the user's token was expired but we were able to connect and get a new one
	newSession := encrypted_session_cache.Session{
		Token: newTokenResponse.AccessToken,
	}
	if err := cache.SaveSession(newSession); err != nil {
		logrus.Warnf("We received a new token from Auth0 but the following error occurred when caching it locally:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		logrus.Warn("If this error isn't corrected, you'll need to log into Kurtosis every time you run it")
		time.Sleep(authWarningPause)
	}
	return newTokenResponse.Scope, nil
}
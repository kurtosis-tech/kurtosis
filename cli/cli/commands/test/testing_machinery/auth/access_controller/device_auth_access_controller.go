/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/access_controller/permissions"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/auth0_authenticators"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/auth0_token_claims"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/test/testing_machinery/auth/session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

type DeviceAuthAccessController struct {
	// Mapping of key_id -> pem_encoded_pubkey_cert for validating tokens
	tokenValidationPubKeys  map[string]string
	sessionCache            session_cache.SessionCache
	deviceCodeAuthenticator auth0_authenticators.DeviceCodeAuthenticator
}

func NewDeviceAuthAccessController(
		tokenValidationPubKeys map[string]string,
		sessionCache session_cache.SessionCache,
		deviceAuthenticator auth0_authenticators.DeviceCodeAuthenticator) *DeviceAuthAccessController {
	return &DeviceAuthAccessController{
		tokenValidationPubKeys:  tokenValidationPubKeys,
		sessionCache:            sessionCache,
		deviceCodeAuthenticator: deviceAuthenticator,
	}
}

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them to authenticate with their username and password

Returns:
	An error if and only if an irrecoverable login error occurred
*/
func (accessController DeviceAuthAccessController) Authenticate() (*permissions.Permissions, error) {
	/*
		NOTE: As of 2020-10-24, we actually don't strictly *need* to encrypt anything on disk because we hardcode the
		 Auth0 public keys used for verifying tokens so unless the user cracks Auth0 and gets the private key, there's
		 no way for a user to forge a token.

		However, this hardcode-public-keys approach becomes much harder if we start doing private key rotation (which would
		 be good security hygiene) because:
		  a) now our code needs to dynamically discover what public keys it should use and
		  b) Kurtosis needs to work even if the developer is offline
		The offline requirement is the real kicker, because it means we need to write the public keys to the developer's local
		 machine and somehow protect it from tampering. This likely means encrypting the data, which means having an encryption
		 key in the code, which would shift the weakpoint to someone decompiling kurtosis-core and discovering the encryption
		 key there.
	*/
	tokenStr, err := getTokenStr(accessController.deviceCodeAuthenticator, accessController.sessionCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	logrus.Debugf("Token str (before expiry check): %v", tokenStr)

	claims, err := parseAndCheckTokenClaims(accessController.tokenValidationPubKeys, tokenStr, accessController.deviceCodeAuthenticator, accessController.sessionCache)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An unrecoverable error occurred checking the token expiration")
	}

	result := parsePermissionsFromClaims(claims)
	return result, nil
}


// ============================ PRIVATE HELPER FUNCTIONS =========================================
/*
Gets the token string, either by reading a valid cache or by prompting the user for their login credentials
*/
func getTokenStr(deviceAuthorizer auth0_authenticators.DeviceCodeAuthenticator, cache session_cache.SessionCache) (string, error) {
	var result string
	session, err := cache.LoadSession()
	if err != nil {
		// We couldn't load any cached session, so the user MUST log in
		logrus.Tracef("The following error occurred loading the session from file: %v", err)
		newToken, err := refreshSession(deviceAuthorizer, cache)
		if err != nil {
			return "", stacktrace.Propagate(err, "No token could be loaded from the cache and an error occurred " +
				"retrieving a new token from Auth0; to continue using Kurtosis you'll need to resolve the error " +
				"and get a new token")
		}
		result = newToken
	} else {
		// We were able to load a session
		result = session.Token
	}

	return result, nil
}

/*
Checks the token expiration and, if the expiration is date is passed but still within the grace period, attempts
	to get a new token

Returns a new claims object if we were able to retrieve the new token
*/
func parseAndCheckTokenClaims(rsaPubKeysPem map[string]string, tokenStr string, deviceAuthorizer auth0_authenticators.DeviceCodeAuthenticator, cache session_cache.SessionCache) (*auth0_token_claims.Auth0TokenClaims, error) {
	claims, err := parseTokenClaims(rsaPubKeysPem, tokenStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing/validating the token claims")
	}

	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt, 0)

	// token is still valid
	if expiration.After(now) {
		return claims, nil
	}

	// At this point, the token is expired so we need to request a new token
	logrus.Infof("Current token expired at '%v'; requesting new token...", expiration)
	newToken, err := refreshSession(deviceAuthorizer, cache)
	if err != nil {
		// The token is expired, we couldn't reach Auth0, and we're beyond the grace period; Kurtosis stops working
		if expiration.Add(tokenExpirationGracePeriod).Before(now) {
			return nil, stacktrace.NewError("Token expired at '%v' which is beyond the " +
				"grace period of %v ago, and we couldn't get a new token from Auth0; to continue using " +
				"Kurtosis you'll need to resolve the error to get a new token",
				expiration,
				tokenExpirationGracePeriod)
		}

		// The token is expired and we couldn't reach Auth0, but we're in the grace period so Kurtosis continues functioning
		expirationExceededAmount := now.Sub(expiration)
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

		return claims, nil
	}

	// If we get here, the current token is expired but we got a new one
	newClaims, err := parseTokenClaims(rsaPubKeysPem, newToken)
	if err != nil {
		return nil, stacktrace.Propagate(err, "We retrieved a new token, but an error occurred parsing/validating the " +
			"token; this is VERY strange that a token we just received from Auth0 is invalid!")
	}

	return newClaims, nil
}

// Attempts to contact Auth0, get a new token, and save the result to the session cache
func refreshSession(deviceAuthorizer auth0_authenticators.DeviceCodeAuthenticator, cache session_cache.SessionCache) (string, error) {
	newToken, err := deviceAuthorizer.AuthorizeDeviceAndAuthenticate()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred retrieving a new token from Auth0")
	}

	newSession := session_cache.Session{
		Token: newToken,
	}
	if err := cache.SaveSession(newSession); err != nil {
		logrus.Warnf("We received a token from Auth0 but the following error occurred when caching it locally:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		logrus.Warn("If this error isn't corrected, you'll need to log into Kurtosis every time you run it")
		time.Sleep(authWarningPause)
	}
	return newToken, nil
}

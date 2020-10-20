/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_authorizer"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/encrypted_session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// Users maybe be out of internet range when their token expires, so we give them a grace period to
	//  get a new token so they don't get surprised by it
	tokenExpirationGracePeriod = 5 * 24 * time.Hour

	// How long we'll pause after displaying auth warnings, to give users a chance to see it
	authWarningPause = 2 * time.Second
)

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them for their creds

Args:
	sessionCacheDirpath: Directory where session cache files will be written

Returns:
	An error if and only if an irrecoverable login error occurred
 */
func RunDeveloperMachineAuthFlow(sessionCacheDirpath string) error {
	cache := encrypted_session_cache.NewEncryptedSessionCache(sessionCacheDirpath, mode)

	tokenStr, err := getTokenStr(cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	claims, err := parseTokenAndGetClaims(tokenStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the token string")
	}

	scope, err := getScopeFromClaimsAndRenewIfNeeded(claims, cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred renewing the token or getting the scope from the token's claims")
	}

	return nil
}

/*
This workflow is for Kurtosis tests running in CI (no device or username).
 */
func RunCIAuthFlow(clientId string, clientSecret string) {
	// TODO
}

/*
	If clientId and clientSecret are non-empty, authorizes based on an OAuth Client ID and Client Secret. ( https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/ )

	If clientId and clientSecret are empty, authorizes a user based on username and password credentials, in addition to device validation.

	In either case, access tokens are cached in a local session.

	Args:
		tokenStorageDirpath: The directory to store the token received from authentication
		clientId: The client ID, for use when running in CI
		clientSecret: The client secret, for use when running in CI

	Returns: Error if
 */
func AuthenticateAndAuthorize(tokenStorageDirpath string, clientId string, clientSecret string) (err error) {

	cache, err := encrypted_session_cache.NewEncryptedSessionCache(tokenStorageDirpath)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to initialize session cache.")
	}

	if (len(clientId) > 0 || len(clientSecret) > 0) && !(len(clientId) > 0 && len(clientSecret) > 0) {
		return false, false, stacktrace.Propagate(err, "If one of clientId or clientSecret are specified, both must be specified. These are only needed when running Kurtosis in CI.")
	}

	isRunningInCI := len(clientId) > 0 && len(clientSecret) > 0

	cachedTokenResponse, alreadyAuthenticated, err := cache.LoadToken()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to load authorization token from session cache at %s", cache.TokenFilePath)
	}

	if alreadyAuthenticated {
		logrus.Debugf("Already authenticated on this device! Access token: %s", cachedTokenResponse.AccessToken)
		return true, cachedTokenResponse.Scope == auth0_authorizer.RequiredScope, nil
	}

	var tokenResponse *auth0_authorizer.TokenResponse
	if isRunningInCI {
		tokenResponse, err = auth0_authorizer.AuthorizeClientCredentials(clientId, clientSecret)
		if err != nil {
			return false, false, stacktrace.Propagate(err, "Failed to authorize client credentials.")
		}
	} else {
		tokenResponse, err = auth0_authorizer.AuthorizeUserDevice()
		if err != nil {
			return false, false, stacktrace.Propagate(err, "Failed to authorize the user and device from auth provider.")
		}
	}

	logrus.Debugf("Access token: %s", tokenResponse.AccessToken)
	err = cache.PersistToken(tokenResponse)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to persist access token to the session cache.")
	}

	return true, tokenResponse.Scope == auth0_authorizer.RequiredScope, nil
}


// ============================== PRIVATE HELPER FUNCTIONS =========================================
func getTokenStr(cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	var result string
	session, err := cache.LoadSession()
	if err != nil {
		// We couldn't load any cached session, so the user MUST log in
		logrus.Debugf("The following error occurred loading the session from file: %v", err)
		tokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
		if err != nil {
			return "", stacktrace.Propagate(err, "An irrecoverable error occurred during Auth0 authorization")
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

func parseTokenAndGetClaims(tokenStr string) (Auth0TokenClaims, error) {
	// This includes validation like expiration date, issuer, etc.
	// See Auth0TokenClaims for more details
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Auth0TokenClaims{},
		// This is the "key extractor" algorithm. We don't return anything here because our
		//  signing algorithm is RSA, which doesn't have secret keys like HMAC which this seems intended for
		func(token *jwt.Token) (interface{}, error) {
			return nil, nil
		},
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

	claims, ok := token.Claims.(Auth0TokenClaims)
	if !ok {
		return Auth0TokenClaims{}, stacktrace.NewError("Could not assert token claims to Auth0 token claims object, indicating an invalid token")
	}

	return claims, nil
}

func getScopeFromClaimsAndRenewIfNeeded(claims Auth0TokenClaims, cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	now := time.Now()
	expiration := time.Unix(claims.ExpiresAt, 0)
	if expiration.Sub(now) >= 0*time.Second {
		return claims.Scope, nil
	}

	// If we've gotten here, it means that the token is beyond the expiration but not beyond the grace period (else
	//  token validation would have failed completely)
	expirationExceededAmount := now.Sub(expiration)
	logrus.Infof("Kurtosis token expired %v ago; attempting to get a new token...", expirationExceededAmount)
	newTokenResponse, err := auth0_authorizer.AuthorizeUserDevice()
	if err != nil {
		logrus.Warnf(
			"WARNING: Your Kurtosis token expired %v ago and we couldn't reach Auth0 to get a new one",
			expirationExceededAmount)
		logrus.Warnf(
			"You will have a grace period of %v from expiration to get a connection to Auth0 before Kurtosis stops working",
			tokenExpirationGracePeriod)
		logrus.Debugf("Token expiration error: %v", err)
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
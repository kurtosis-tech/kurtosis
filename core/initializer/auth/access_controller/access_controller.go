/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authorizer"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_authorizers"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_constants"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/auth0_token_claims"
	"github.com/kurtosis-tech/kurtosis/initializer/auth/encrypted_session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
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

	// Header and footer to attach to base64-encoded key data that we receive from Auth0
	pubKeyHeader = "-----BEGIN CERTIFICATE-----"
	pubKeyFooter = "-----END CERTIFICATE-----"
)

type AccessController struct {
	// Mapping of key_id -> base64_encoded_pubkey for validating tokens
	tokenValidationPubKeys map[string]string
	clientCredsAuthorizer auth0_authorizers.ClientCredentialsAuthorizer
	deviceAuthorizer auth0_authorizers.DeviceAuthorizer
}

func NewAccessController(
		tokenValidationPubKeys map[string]string,
		clientCredsAuthorizer auth0_authorizers.ClientCredentialsAuthorizer) *AccessController {
	return &AccessController{
		tokenValidationPubKeys: tokenValidationPubKeys,
		clientCredsAuthorizer: clientCredsAuthorizer,
	}
}

/*
Used for a developer running Kurtosis on their local machine. This will:

1) Check if they have a valid session cached locally that's still valid and, if not
2) Prompt them for their username and password

Args:
	sessionCacheFilepath: Filepath to store the encrypted session cache at

Returns:
	An error if and only if an irrecoverable login error occurred
 */
func (accessController AccessController) RunDeveloperMachineAuthFlow(sessionCacheFilepath string) error {
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
	cache := encrypted_session_cache.NewEncryptedSessionCache(sessionCacheFilepath, sessionCacheFilePerms)

	tokenStr, err := getTokenStr(accessController.deviceAuthorizer, cache)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the token string")
	}

	claims, err := parseAndCheckTokenClaims(tokenStr, accessController.deviceAuthorizer, cache)
	if err != nil {
		return stacktrace.Propagate(err, "An unrecoverable error occurred checking the token expiration")
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

/*
This workflow is for authenticating and authorizing Kurtosis tests running in CI (no device or username).
	See also: https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/
 */
func (accessController AccessController) RunCIAuthFlow(clientId string, clientSecret_DO_NOT_LOG string) error {
	tokenResponse, err := accessController.clientCredsAuthorizer.AuthorizeClientCredentials(clientId, clientSecret_DO_NOT_LOG)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred authenticating with the client ID & secret")
	}

	claims, err := parseTokenClaims(tokenResponse.AccessToken)
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
func getTokenStr(deviceAuthorizer auth0_authorizers.DeviceAuthorizer, cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
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

Returns a new claims object if we were able to
 */
func parseAndCheckTokenClaims(tokenStr string, deviceAuthorizer auth0_authorizers.DeviceAuthorizer, cache *encrypted_session_cache.EncryptedSessionCache) (*auth0_token_claims.Auth0TokenClaims, error) {
	claims, err := parseTokenClaims(tokenStr)
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
	newClaims, err := parseTokenClaims(newToken)
	if err != nil {
		return nil, stacktrace.Propagate(err, "We retrieved a new token, but an error occurred parsing/validating the " +
			"token; this is VERY strange that a token we just received from Auth0 is invalid!")
	}

	return newClaims, nil
}

// Parses a token string, validates the claims, and returns the claims object
func parseTokenClaims(tokenStr string) (*auth0_token_claims.Auth0TokenClaims, error) {
	token, err := new(jwt.Parser).ParseWithClaims(
		tokenStr,
		&auth0_token_claims.Auth0TokenClaims{},
		getPubKeyFromKurtosisToken,
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

// jwt-go requires that you have a function which uses the token to get the public key for validating the tokne
// This is our method for doing this with Kurtosis tokens
func getPubKeyFromKurtosisToken(token *jwt.Token) (interface{}, error) {
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

	keyBase64, found := auth0_constants.RsaPublicKeyBase64[keyId]
	if !found {
		return nil, stacktrace.NewError("No public RSA key found corresponding to key ID from token '%v'", keyId)
	}
	keyStr := pubKeyHeader + "\n" + keyBase64 + "\n" + pubKeyFooter

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(keyStr))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the public key base64 for key ID '%v'; this is a code bug", keyId)
	}

	return pubKey, nil
}

// Attempts to contact Auth0, get a new token, and save the result to the session cache
func refreshSession(deviceAuthorizer auth0_authorizers.DeviceAuthorizer, cache *encrypted_session_cache.EncryptedSessionCache) (string, error) {
	newTokenResponse, err := deviceAuthorizer.AuthorizeUserDevice()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred retrieving a new token from Auth0")
	}
	newToken := newTokenResponse.AccessToken

	newSession := encrypted_session_cache.Session{
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
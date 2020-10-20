/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_authorizer"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
)

/*
	If clientId and clientSecret are non-empty, authorizes based on an OAuth Client ID and Client Secret. ( https://www.oauth.com/oauth2-servers/access-tokens/client-credentials/ )
	This workflow is for Kurtosis tests running in CI (no device or username).

	If clientId and clientSecret are empty, authorizes a user based on username and password credentials, in addition to device validation.

	In either case, access tokens are cached in a local session.

	Args:
		tokenStorageDirpath: The directory to store the token received from authentication
		clientId: The client ID, for use when running in CI
		clientSecret: The client secret, for use when running in CI

	Returns:
		authenticated: Bool indicating if the user exists in the system.
		authorized: Bool indicating if the user is authorized to run Kurtosis.
 */
func AuthenticateAndAuthorize(tokenStorageDirpath string, clientId string, clientSecret string) (authenticated bool, authorized bool, err error) {
	// TODO Switch to https://github.com/gorilla/sessions for token storage
	//  per https://auth0.com/docs/quickstart/webapp/golang ???
	cache, err := session_cache.NewSessionCache(tokenStorageDirpath)
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

func validateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, keyFunction)

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["foo"], claims["nbf"])
	} else {
		fmt.Println(err)
	}
	return true, nil
}

func keyFunction(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, stacktrace.NewError("Unexpected signing method: %v", token.Header["alg"])
	}

	// TODO TODO TODO THIS MUST RETURN BYTE ARRAY OF RSA PUBLIC KEY
	return []byte("Hello."), nil
}
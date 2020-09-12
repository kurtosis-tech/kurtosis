/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/session_cache"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
	If ciLicense is non-empty, bypasses authentication TODO TODO TODO implement CI machine to machine auth.
	If ciLicense is empty, run a user and device-based authentication flow, prompting the user to login.
	Returns authenticated to indicate if the user exists in the system.
	Returns authorized if the user is authorized to run Kurtosis.
 */
func AuthenticateAndAuthorize(ciLicense string) (authenticated bool, authorized bool, err error) {
	cache, err := session_cache.NewSessionCache()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to initialize session cache.")
	}
	if len(ciLicense) > 0 {
		// TODO TODO TODO Implement machine-to-machine auth flow to actually do auth for CI workflows https://auth0.com/docs/applications/set-up-an-application/register-machine-to-machine-applications
		return true, true, nil
	}
	tokenResponse, alreadyAuthenticated, err := cache.LoadToken()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to load authorization token from session cache at %s", cache.TokenFilePath)
	}
	if alreadyAuthenticated {
		logrus.Debugf("Already authenticated on this device! Access token: %s", tokenResponse.AccessToken)
		return true, tokenResponse.Scope == auth0.RequiredScope, nil
	}
	tokenResponse, err = auth0.AuthorizeUserDevice()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to authorize the user and device from auth provider.")
	}
	logrus.Debugf("Access token: %s", tokenResponse.AccessToken)
	err = cache.PersistToken(tokenResponse)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "Failed to persist access token to the session cache.")
	}

	return true, tokenResponse.Scope == auth0.RequiredScope, nil
}

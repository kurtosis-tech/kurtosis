/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	requiredScopeToExecute = "execute:kurtosis-core"
)

/*
	If ciLicense is non-empty, bypasses authentication TODO TODO TODO implement CI machine to machine auth.
	If ciLicense is empty, run a user and device-based authentication flow, prompting the user to login.
	Returns authenticated to indicate if the user exists in the system.
	Returns authorized if the user is authorized to run Kurtosis.
 */
func AuthenticateAndAuthorize(ciLicense string) (authenticated bool, authorized bool, err error) {
	if len(ciLicense) > 0 {
		// TODO TODO TODO Implement machine-to-machine auth flow to actually do auth for CI workflows https://auth0.com/docs/applications/set-up-an-application/register-machine-to-machine-applications
		return true, true, nil
	}
	tokenResponse, alreadyAuthenticated, err := loadToken()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "")
	}
	if alreadyAuthenticated {
		logrus.Debugf("Already authenticated on this device! Access token: %s", tokenResponse.AccessToken)
		return true, tokenResponse.Scope == requiredScopeToExecute, nil
	}
	tokenResponse, err = authorizeUserDevice()
	if err != nil {
		return false, false, stacktrace.Propagate(err, "")
	}
	logrus.Debugf("Access token: %s", tokenResponse.AccessToken)
	err = persistToken(tokenResponse)
	if err != nil {
		return false, false, stacktrace.Propagate(err, "")
	}
	return true, tokenResponse.Scope == requiredScopeToExecute, nil
}

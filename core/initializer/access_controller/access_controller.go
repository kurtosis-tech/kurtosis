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
	Verifies that the license is a credential for a registered user,
	Returns true if the license associated with the user has ever
	been registered in Kurtosis, even if that user has no permissions.
 */
func AuthenticateAndAuthorize(ciLicense string) (authenticated bool, authorized bool, err error) {
	if len(ciLicense) > 0 {
		// TODO TODO TODO Implement machine-to-machine auth flow to actually do auth for CI workflows
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

/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

/*
	Verifies that the license is a credential for a registered user,
	Returns true if the license associated with the user has ever
	been registered in Kurtosis, even if that user has no permissions.
 */
func Authenticate() (authenticated bool, err error) {
	tokenResponse, alreadyAuthenticated, err := loadToken()
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	if alreadyAuthenticated {
		logrus.Debugf("Already authenticated on this device! Access token: %s", tokenResponse.AccessToken)
		return true, nil
	}
	tokenResponse, err = authorizeUserDevice()
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
	logrus.Debugf("Access token: %s", tokenResponse.AccessToken)
	err = persistToken(tokenResponse)
	if err != nil {
		return false, stacktrace.Propagate(err, "")
	}
}

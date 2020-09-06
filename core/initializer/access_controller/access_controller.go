/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import "github.com/palantir/stacktrace"

const (
	freeTrialLicense = "free-trial"
	expiredTrialLicense = "expired-trial"
	paidLicense = "paid-license"
)

/*
	Verifies that the license is a credential for a registered user,
	Returns true if the license associated with the user has ever
	been registered in Kurtosis, even if that user has no permissions.
 */
func AuthenticateLicense(license string) (bool, error) {
	switch license {
	case freeTrialLicense, paidLicense, expiredTrialLicense:
		return true, nil
	default:
		return false, nil
	}
}

/*
	Verifies that the user associated with the license is authorized
	to run kurtosis (either has a free trial or has purchased a full license).
*/
func AuthorizeLicense(license string) (bool, error) {
	switch license {
	case freeTrialLicense, paidLicense:
		return true, nil
	case expiredTrialLicense:
		return false, nil
	default:
		return false, stacktrace.NewError("Failed to authorize license.")
	}
}

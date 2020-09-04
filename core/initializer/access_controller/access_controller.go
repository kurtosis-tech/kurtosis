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

func AuthenticateLicense(license string) (bool, error) {
	switch license {
	case freeTrialLicense, paidLicense, expiredTrialLicense:
		return true, nil
	default:
		return false, nil
	}
}

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

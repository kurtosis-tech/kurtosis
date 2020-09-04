/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

const (
	freeTrialLicense = "free-trial"
	expiredTrialLicense = "expired-trial"
	paidLicense = "paid-license"
)

func AuthenticateLicense(license string) bool {
	switch license {
	case freeTrialLicense, paidLicense, expiredTrialLicense:
		return true
	default:
		return false
	}
}

func AuthorizeLicense(license string) bool {
	switch license {
	case freeTrialLicense, paidLicense:
		return true
	case expiredTrialLicense:
		return false
	}
}

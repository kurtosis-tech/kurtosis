/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_ids

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"regexp"
	"time"
)

const (
	EnclaveIdMaxLength            = 63 // If changing this, change also the regexp AllowedEnclaveIdCharsRegexStr below
	AllowedEnclaveIdCharsRegexStr = `^[-A-Za-z0-9.]{1,63}$`

	kurtosisPrefix = "kt"
	// YYYY-MM-DDTHH-MM-SS-sss
	executionTimestampFormat = "2006-01-02t15-04-05-000"
)

func GenerateNewEnclaveID() string {
	return fmt.Sprintf(
		"%v%v",
		kurtosisPrefix,
		time.Now().Format(executionTimestampFormat),
	)
}

func ValidateEnclaveId(enclaveIdStr string) error {
	// TODO Push down into MetricsReportingKurtosisBackend
	validEnclaveId, err := regexp.Match(AllowedEnclaveIdCharsRegexStr, []byte(enclaveIdStr))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave ID '%v' matches allowed enclave ID regex '%v'",
			enclaveIdStr,
			AllowedEnclaveIdCharsRegexStr,
		)
	}
	if !validEnclaveId {
		return stacktrace.NewError(
			"Enclave ID '%v' doesn't match allowed enclave ID regex '%v'",
			enclaveIdStr,
			AllowedEnclaveIdCharsRegexStr,
		)
	}
	return nil
}

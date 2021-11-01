/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package object_labels_providers

import "github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_object_labels"

// TODO This magic function that Kurtosis devs just have to remember to use kinda sucks; we want to force them to use this
// Every Kurtosis object must have the Kurtosis app ID & value
func getLabelsForKurtosisObject() map[string]string {
	return map[string]string{
		enclave_object_labels.AppIDLabel: enclave_object_labels.AppIDValue,
	}
}

/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kurtosis_backend_config

type KubernetesBackendConfig struct {
	StorageClass string `json:"storage-class"`
	EnclaveSizeInGigabytes uint `json:"enclave-size-in-gigabytes"`
}

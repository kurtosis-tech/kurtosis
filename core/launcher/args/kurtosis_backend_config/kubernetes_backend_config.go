/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kurtosis_backend_config

type KubernetesBackendConfig struct {
	StorageClass string `json:"storageClass"`


	// TODO Remove this and replace wtih FilesArtifactExpansionVolumeSizeInMegabytes, because the API container
	//  actually doesn't need enclave data volume size (because it doesn't create enclaves)
	EnclaveSizeInMegabytes uint `json:"enclaveSizeInMegabytes"`
}

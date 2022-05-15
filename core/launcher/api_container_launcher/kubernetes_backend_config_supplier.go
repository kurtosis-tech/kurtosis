/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args/kurtosis_backend_config"
)

type KubernetesBackendConfigSupplier struct {
	storageClass string
	enclaveSizeInGigabytes uint
}

func NewKubernetesKurtosisBackendConfigSupplier(storageClass string, enclaveSizeInGigabytes uint) KubernetesBackendConfigSupplier {
	return KubernetesBackendConfigSupplier{
		storageClass: storageClass,
		enclaveSizeInGigabytes: enclaveSizeInGigabytes,
	}
}

func (backendConfigSupplier KubernetesBackendConfigSupplier) getKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Kubernetes, kurtosis_backend_config.KubernetesBackendConfig{
		StorageClass: backendConfigSupplier.storageClass,
		EnclaveSizeInGigabytes: backendConfigSupplier.enclaveSizeInGigabytes,
	}
}


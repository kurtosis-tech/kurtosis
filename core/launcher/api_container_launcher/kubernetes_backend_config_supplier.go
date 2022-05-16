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
	enclaveSizeInMegabytes uint
}

func NewKubernetesKurtosisBackendConfigSupplier(storageClass string, enclaveSizeInMegabytes uint) KubernetesBackendConfigSupplier {
	return KubernetesBackendConfigSupplier{
		storageClass: storageClass,
		enclaveSizeInMegabytes: enclaveSizeInMegabytes,
	}
}

func (backendConfigSupplier KubernetesBackendConfigSupplier) getKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Kubernetes, kurtosis_backend_config.KubernetesBackendConfig{
		StorageClass: backendConfigSupplier.storageClass,
		EnclaveSizeInMegabytes: backendConfigSupplier.enclaveSizeInMegabytes,
	}
}


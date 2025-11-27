/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"github.com/dzobbe/PoTE-kurtosis/core/launcher/args"
	"github.com/dzobbe/PoTE-kurtosis/core/launcher/args/kurtosis_backend_config"
)

type KubernetesBackendConfigSupplier struct {
	storageClass string
}

func NewKubernetesKurtosisBackendConfigSupplier(storageClass string) KubernetesBackendConfigSupplier {
	return KubernetesBackendConfigSupplier{
		storageClass: storageClass,
	}
}

func (backendConfigSupplier KubernetesBackendConfigSupplier) getKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Kubernetes, kurtosis_backend_config.KubernetesBackendConfig{StorageClass: backendConfigSupplier.storageClass}
}

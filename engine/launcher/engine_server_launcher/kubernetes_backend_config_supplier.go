/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_backend_config"
)

type KubernetesBackendConfigSupplier struct {
	kubernetesBackendConfig kurtosis_backend_config.KubernetesBackendConfig
}

func NewKubernetesKurtosisBackendConfigSupplier(storageClass string, enclaveVolumeSizeInGB uint) KubernetesBackendConfigSupplier {
	kubernetesBackendConfig := kurtosis_backend_config.KubernetesBackendConfig{
		storageClass,
		enclaveVolumeSizeInGB,
	}
	return KubernetesBackendConfigSupplier{kubernetesBackendConfig: kubernetesBackendConfig}
}

func (backendConfigSupplier *KubernetesBackendConfigSupplier) getKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Kubernetes, backendConfigSupplier.kubernetesBackendConfig
}


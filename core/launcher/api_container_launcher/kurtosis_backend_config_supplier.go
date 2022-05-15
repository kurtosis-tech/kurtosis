/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args/kurtosis_backend_config"
)

type KurtosisBackendConfigSupplier interface {
	// Private because only the launcher should call it
	getKurtosisBackendConfig() (args.KurtosisBackendType, interface{})
}

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

func (backendConfigSupplier *KubernetesBackendConfigSupplier) getKurtosisBackendConfig () (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Kubernetes, backendConfigSupplier.kubernetesBackendConfig
}

type DockerBackendConfigSupplier struct {
	dockerBackendConfig kurtosis_backend_config.DockerBackendConfig
}

func NewDockerKurtosisBackendConfigSupplier() DockerBackendConfigSupplier {
	dockerBackendConfig := kurtosis_backend_config.DockerBackendConfig{}
	return DockerBackendConfigSupplier{dockerBackendConfig: dockerBackendConfig}
}

func (backendConfigSupplier *DockerBackendConfigSupplier) getKurtosisBackendConfig () (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Docker, backendConfigSupplier.dockerBackendConfig
}

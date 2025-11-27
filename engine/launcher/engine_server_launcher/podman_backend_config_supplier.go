/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */
package engine_server_launcher

import (
	"github.com/dzobbe/PoTE-kurtosis/engine/launcher/args"
	"github.com/dzobbe/PoTE-kurtosis/engine/launcher/args/kurtosis_backend_config"
)

type PodmanBackendConfigSupplier struct {
}

func NewPodmanKurtosisBackendConfigSupplier() PodmanBackendConfigSupplier {
	return PodmanBackendConfigSupplier{}
}

func (backendConfigSupplier PodmanBackendConfigSupplier) getKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	dockerBackendConfig := kurtosis_backend_config.DockerBackendConfig{}
	return args.KurtosisBackendType_Podman, dockerBackendConfig
}

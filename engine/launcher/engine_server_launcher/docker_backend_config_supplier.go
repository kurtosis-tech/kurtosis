/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args/kurtosis_backend_config"
)

type DockerBackendConfigSupplier struct {
	dockerBackendConfig kurtosis_backend_config.DockerBackendConfig
}

func NewDockerKurtosisBackendConfigSupplier() DockerBackendConfigSupplier {
	dockerBackendConfig := kurtosis_backend_config.DockerBackendConfig{}
	return DockerBackendConfigSupplier{dockerBackendConfig: dockerBackendConfig}
}

func (backendConfigSupplier *DockerBackendConfigSupplier) GetKurtosisBackendConfig() (args.KurtosisBackendType, interface{}) {
	return args.KurtosisBackendType_Docker, backendConfigSupplier.dockerBackendConfig
}


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
}

func NewDockerKurtosisBackendConfigSupplier() DockerBackendConfigSupplier {
	return DockerBackendConfigSupplier{}
}

func (backendConfigSupplier DockerBackendConfigSupplier) getKurtosisBackendConfig () (args.KurtosisBackendType, interface{}) {
	dockerBackendConfig := kurtosis_backend_config.DockerBackendConfig{}
	return args.KurtosisBackendType_Docker, dockerBackendConfig
}


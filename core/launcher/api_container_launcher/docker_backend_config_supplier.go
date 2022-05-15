/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args/kurtosis_backend_config"
)

type DockerBackendConfigSupplier struct {
}

func NewDockerKurtosisBackendConfigSupplier() DockerBackendConfigSupplier {
}

func (backendConfigSupplier DockerBackendConfigSupplier) getKurtosisBackendConfig () (args.KurtosisBackendType, interface{}) {
	dockerBackendConfig := kurtosis_backend_config.DockerBackendConfig{}
	return args.KurtosisBackendType_Docker, dockerBackendConfig
}


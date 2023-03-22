/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/remote_context_backend"
	contexts_config_store_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	contexts_config_store_generated_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
)

type KurtosisContextSupplier func() (*contexts_config_store_generated_api.KurtosisContext, error)

type KurtosisRemoteBackendConfigSupplier struct {
	kurtosisContextSupplier KurtosisContextSupplier
}

func NewKurtosisRemoteBackendConfigSupplier(kurtosisContextSupplier KurtosisContextSupplier) *KurtosisRemoteBackendConfigSupplier {
	return &KurtosisRemoteBackendConfigSupplier{
		kurtosisContextSupplier: kurtosisContextSupplier,
	}
}

func (supplier KurtosisRemoteBackendConfigSupplier) GetOptionalRemoteConfig() (*remote_context_backend.KurtosisRemoteBackendConfig, error) {
	kurtosisContext, err := supplier.kurtosisContextSupplier()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error retrieving Kurtosis context object")
	}

	convertToRemoteBackendConfigVisitor := contexts_config_store_api.KurtosisContextVisitor[remote_context_backend.KurtosisRemoteBackendConfig]{
		VisitLocalOnlyContextV0: func(localContext *contexts_config_store_generated_api.LocalOnlyContextV0) (*remote_context_backend.KurtosisRemoteBackendConfig, error) {
			// context is local, no remote backend config
			return nil, nil
		},
		VisitRemoteContextV0: func(remoteContext *contexts_config_store_generated_api.RemoteContextV0) (*remote_context_backend.KurtosisRemoteBackendConfig, error) {
			var tlsConfig *remote_context_backend.KurtosisBackendTlsConfig
			if remoteContext.GetTlsConfig() != nil {
				tlsConfig = &remote_context_backend.KurtosisBackendTlsConfig{
					Ca:         remoteContext.GetTlsConfig().GetCertificateAuthority(),
					ClientCert: remoteContext.GetTlsConfig().GetClientCertificate(),
					ClientKey:  remoteContext.GetTlsConfig().GetClientKey(),
				}
			}
			return &remote_context_backend.KurtosisRemoteBackendConfig{
				Host: remoteContext.GetHost(),
				Port: remoteContext.GetKurtosisBackendPort(),
				Tls:  tlsConfig,
			}, nil
		},
	}
	remoteBackendConfigMaybe, err := contexts_config_store_api.Visit[remote_context_backend.KurtosisRemoteBackendConfig](kurtosisContext, convertToRemoteBackendConfigVisitor)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to generate remote backend configuration")
	}
	return remoteBackendConfigMaybe, nil
}

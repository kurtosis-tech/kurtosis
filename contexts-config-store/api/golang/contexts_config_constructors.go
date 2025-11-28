package golang

import "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"

func NewContextUuid(uuid string) *generated.ContextUuid {
	return &generated.ContextUuid{
		Value: uuid,
	}
}

func NewKurtosisContextsConfig(currentContextUuid *generated.ContextUuid, contexts ...*generated.KurtosisContext) *generated.KurtosisContextsConfig {
	return &generated.KurtosisContextsConfig{
		CurrentContextUuid: currentContextUuid,
		Contexts:           contexts,
	}
}

func NewLocalOnlyContext(uuid *generated.ContextUuid, name string) *generated.KurtosisContext {
	return &generated.KurtosisContext{
		Uuid: uuid,
		Name: name,
		KurtosisContextInfo: &generated.KurtosisContext_LocalOnlyContextV0{
			LocalOnlyContextV0: &generated.LocalOnlyContextV0{},
		},
	}
}

func NewRemoteV0Context(
	uuid *generated.ContextUuid,
	name string,
	host string,
	remotePortalPort uint32,
	kurtosisBackendPort uint32,
	tunnelPort uint32,
	tlsConfig *generated.TlsConfig,
	envVars *string,
	cloudUserId *string,
	cloudInstanceId *string,
) *generated.KurtosisContext {
	return &generated.KurtosisContext{
		Uuid: uuid,
		Name: name,
		KurtosisContextInfo: &generated.KurtosisContext_RemoteContextV0{
			RemoteContextV0: &generated.RemoteContextV0{
				Host:                host,
				RemotePortalPort:    remotePortalPort,
				KurtosisBackendPort: kurtosisBackendPort,
				TunnelPort:          tunnelPort,
				TlsConfig:           tlsConfig,
				EnvVars:             envVars,
				CloudUserId:         cloudUserId,
				CloudInstanceId:     cloudInstanceId,
			},
		},
	}
}

package configs

import (
	"encoding/json"
	"fmt"

	contexts_config_store_generated_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/stacktrace"
)

// TODO: Move to the backend interface dir to be more generic
const (
	urlScheme = "tcp"
)

var (
	NoRemoteBackendConfig *KurtosisRemoteBackendConfig = nil
)

// Backend agnostic remote backend config which can be used by the Docker or Kubernetes backend
type KurtosisRemoteBackendConfig struct {
	Endpoint string                    `json:"endpoint"`
	Tls      *KurtosisBackendTlsConfig `json:"tlsConfig,omitempty"`
}

type KurtosisBackendTlsConfig struct {
	Ca         []byte `json:"certificateAuthority"`
	ClientCert []byte `json:"clientCertificate"`
	ClientKey  []byte `json:"clientKey"`
}

func NewRemoteBackendConfigFromRemoteContext(
	remoteContext *contexts_config_store_generated_api.RemoteContextV0,
) *KurtosisRemoteBackendConfig {
	var tlsConfig *KurtosisBackendTlsConfig
	if remoteContext.GetTlsConfig() != nil {
		tlsConfig = &KurtosisBackendTlsConfig{
			Ca:         remoteContext.GetTlsConfig().GetCertificateAuthority(),
			ClientCert: remoteContext.GetTlsConfig().GetClientCertificate(),
			ClientKey:  remoteContext.GetTlsConfig().GetClientKey(),
		}
	}
	endpoint := fmt.Sprintf("%s://%s:%d", urlScheme, remoteContext.GetHost(), remoteContext.GetKurtosisBackendPort())
	return &KurtosisRemoteBackendConfig{
		Endpoint: endpoint,
		Tls:      tlsConfig,
	}
}

// Used to parse the remote backend config on the bastion host
func NewRemoteBackendConfigFromJSON(
	data []byte,
) (*KurtosisRemoteBackendConfig, error) {
	var remoteBackendConfig KurtosisRemoteBackendConfig
	if err := json.Unmarshal(data, &remoteBackendConfig); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to unmarshal docker remote backend config")
	}

	return &remoteBackendConfig, nil
}

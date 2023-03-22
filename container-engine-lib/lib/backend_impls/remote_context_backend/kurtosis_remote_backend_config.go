package remote_context_backend

type KurtosisRemoteBackendConfig struct {
	Host string `json:"host"`
	Port uint32 `json:"port"`

	Tls *KurtosisBackendTlsConfig `json:"tls,omitempty"`
}

type KurtosisBackendTlsConfig struct {
	Ca         []byte `json:"ca"`
	ClientCert []byte `json:"cert"`
	ClientKey  []byte `json:"key"`
}

package config_version

//go:generate go run github.com/dmarkham/enumer -type=ConfigVersion
type ConfigVersion uint

const (
	// To add new values, just add a new version to the end WITHOUT WHITESPACE
	ConfigVersion_v0 ConfigVersion = iota
	ConfigVersion_v1
	ConfigVersion_v2 // Fixed a typo in Kubernetes config, `enclave-size-in-Megabytes` -> `enclave-size-in-megabytes`
	ConfigVersion_v3 // adds logs aggregator config
	ConfigVersion_v4 // adds engine-node-name to KubernetesClusterConfig
	ConfigVersion_v5 // adds GraflokiConfig to KurtosisClusterConfig
)

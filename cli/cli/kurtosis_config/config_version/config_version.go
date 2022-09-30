package config_version


//go:generate go run github.com/dmarkham/enumer -type=ConfigVersion
type ConfigVersion uint
const (
	// To add new values, just add a new version to the end WITHOUT WHITESPACE
	ConfigVersion_v0 ConfigVersion = iota
	ConfigVersion_v1
	ConfigVersion_v2	// Fixed a typo in Kubernetes config, `enclave-size-in-Megabytes` -> `enclave-size-in-megabytes`
)

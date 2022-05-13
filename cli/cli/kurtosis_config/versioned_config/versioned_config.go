package versioned_config

import "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"

// VersionedKurtosisConfig contains the version property, and is intended to be embedded inside all the
// other Kurtosis config objects
type VersionedKurtosisConfig struct {
	ConfigVersion config_version.ConfigVersion `yaml:"config-version"`
}

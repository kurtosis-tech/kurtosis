package v2

import "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult

type KurtosisConfigV2 struct {
	// vvvvvvvvv Every new Kurtosis config version must have this key vvvvvvvv
	ConfigVersion config_version.ConfigVersion `yaml:"config-version"`
	// ^^^^^^^^^ Every new Kurtosis config version must have this key ^^^^^^^^

	ShouldSendMetrics *bool                              `yaml:"should-send-metrics,omitempty"`
	KurtosisClusters map[string]*KurtosisClusterConfigV2 `yaml:"kurtosis-clusters,omitempty"`
}

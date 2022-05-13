package v1

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/versioned_config"
)

// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult

type KurtosisConfigV1 struct {
	versioned_config.VersionedKurtosisConfig

	//We set public fields because YAML marshalling needs it on this way
	//All fields should be pointers, that way we can enforce required fields
	//by detecting nil pointers.
	ConfigVersion *config_version.ConfigVersion `yaml:"config-version,omitempty"`
	ShouldSendMetrics *bool                               `yaml:"should-send-metrics,omitempty"`
	KurtosisClusters *map[string]*KurtosisClusterConfigV1 `yaml:"kurtosis-clusters,omitempty"`
}

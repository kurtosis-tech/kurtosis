package overrides_deserializers

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v0"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v1"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v2"
	"github.com/kurtosis-tech/stacktrace"
)

type configOverridesDeserializer func(configFileBytes []byte) (interface{}, error)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
// Adding a new version here is as simple as:
//   1) copy-pasting a version block
//   2) changing the key to your new config version
//   3) changing the struct that's being deserialized into
// We keep these sorted in REVERSE chronological order so you don't need to scroll
//  to the bottom each time
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
var AllConfigOverridesDeserializers = map[config_version.ConfigVersion]configOverridesDeserializer{
	config_version.ConfigVersion_v2: func(configFileBytes []byte) (interface{}, error) {
		overrides := &v2.KurtosisConfigV2{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(configFileBytes))
		}
		return overrides, nil
	},
	config_version.ConfigVersion_v1: func(configFileBytes []byte) (interface{}, error) {
		overrides := &v1.KurtosisConfigV1{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(configFileBytes))
		}
		return overrides, nil
	},
	config_version.ConfigVersion_v0: func(configFileBytes []byte) (interface{}, error) {
		overrides := &v0.KurtosisConfigV0{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(configFileBytes))
		}
		return overrides, nil
	},
}
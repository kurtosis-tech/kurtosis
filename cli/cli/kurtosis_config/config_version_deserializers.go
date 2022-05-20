package kurtosis_config

import "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"

type configVersionDeserializer func(configFileBytes []byte) (interface{}, error)

// Completeness enforced in a unit test
var configDeserializationFuncs = map[config_version.ConfigVersion]configVersionDeserializer{
	config_version.ConfigVersion_v0: configVersionDeserializer(configFileBytes []byte),
}


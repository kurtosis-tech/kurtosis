package v0

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/stacktrace"
)

// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult
type KurtosisConfigV0 struct {
	//We set public fields because YAML marshalling needs it on this way
	//All fields should be pointers, that way we can enforce required fields
	//by detecting nil pointers.
	ConfigVersion *config_version.ConfigVersion `yaml:"config-version,omitempty"`
	ShouldSendMetrics *bool `yaml:"should-send-metrics,omitempty"`
}

func NewKurtosisConfigV0(doesUserAcceptSendingMetrics *bool) *KurtosisConfigV0 {
	configVersion := config_version.ConfigVersion_v0
	return &KurtosisConfigV0{
		ConfigVersion: &configVersion,
		ShouldSendMetrics: doesUserAcceptSendingMetrics,
	}
}

func (kurtosisConfigV0 *KurtosisConfigV0) Validate() error {
	if kurtosisConfigV0.ShouldSendMetrics == nil {
		return stacktrace.NewError("ShouldSendMetrics field of Kurtosis Config v0 is nil, when it should be true or false.")
	}
	return nil
}
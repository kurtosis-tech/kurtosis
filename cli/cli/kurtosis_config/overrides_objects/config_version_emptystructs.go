package overrides_objects

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/config_version"
	v0 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v0"
	v1 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v1"
	v2 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v2"
	v3 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	v4 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v4"
	v5 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v5"
	v6 "github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_config/overrides_objects/v6"
)

/*
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
When you add a new config version, simply add a new entry here that instantiates a pointer to the emptystruct
of the config version struct.

For an explanation, see the docs on TestKurtosisConfigIsUsingLatestConfigStruct.
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
*/

var AllConfigVersionEmptyStructs = map[config_version.ConfigVersion]interface{}{
	config_version.ConfigVersion_v6: &v6.KurtosisConfigV6{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
	},
	config_version.ConfigVersion_v5: &v5.KurtosisConfigV5{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
	},
	config_version.ConfigVersion_v4: &v4.KurtosisConfigV4{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
	},
	config_version.ConfigVersion_v3: &v3.KurtosisConfigV3{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
	},
	config_version.ConfigVersion_v2: &v2.KurtosisConfigV2{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
	},
	config_version.ConfigVersion_v1: &v1.KurtosisConfigV1{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
	},
	config_version.ConfigVersion_v0: &v0.KurtosisConfigV0{
		ShouldSendMetrics: nil,
	},
}

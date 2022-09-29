package overrides_objects

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	v0 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v0"
	v1 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v1"
	v2 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v2"
)

/*
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
When you add a new config version, simply add a new entry here that instantiates a pointer to the emptystruct
of the config version struct.

For an explanation, see the docs on TestKurtosisConfigIsUsingLatestConfigStruct.
>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
 */

var AllConfigVersionEmptyStructs = map[config_version.ConfigVersion]interface{}{
	config_version.ConfigVersion_v2: &v2.KurtosisConfigV2{},
	config_version.ConfigVersion_v1: &v1.KurtosisConfigV1{},
	config_version.ConfigVersion_v0: &v0.KurtosisConfigV0{},
}

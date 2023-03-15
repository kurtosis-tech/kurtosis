package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"testing"
)

type readyConditionsTestCase struct {
	*testing.T
}

func newReadyConditionsTestCase(t *testing.T) *readyConditionsTestCase {
	return &readyConditionsTestCase{
		T: t,
	}
}

func (t *readyConditionsTestCase) GetId() string {
	return service_config.ReadyConditionsTypeName
}

func (t *readyConditionsTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return service_config.NewReadyConditionsType()
}

/*
get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
	)

ready_conditions = ReadyConditions(

        recipe=get_recipe,
		field="code",
		assertion="==",
		target_value=%v,
		interval="1s",
		timeout="3s"
    )
*/

func (t *serviceConfigFullTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s(%s=%q, %s=%q), %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)",
		service_config.ReadyConditionsTypeName,
		service_config.RecipeAttr,
		recipe.GetHttpRecipeTypeName,
		/*service_config.PortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}", TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol),
		service_config.PublicPortsAttr, fmt.Sprintf("{%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}", TestPublicPortId, TestPublicPortNumber, TestPublicPortProtocolStr, TestPublicApplicationProtocol),
		service_config.FilesAttr, fmt.Sprintf("{%q: %q, %q: %q}", TestFilesArtifactPath1, TestFilesArtifactName1, TestFilesArtifactPath2, TestFilesArtifactName2),
		service_config.EntrypointAttr, fmt.Sprintf("[%q, %q]", TestEntryPointSlice[0], TestEntryPointSlice[1]),
		service_config.CmdAttr, fmt.Sprintf("[%q, %q, %q]", TestCmdSlice[0], TestCmdSlice[1], TestCmdSlice[2]),
		service_config.EnvVarsAttr, fmt.Sprintf("{%q: %q, %q: %q}", TestEnvVarName1, TestEnvVarValue1, TestEnvVarName2, TestEnvVarValue2),
		service_config.PrivateIpAddressPlaceholderAttr, TestPrivateIPAddressPlaceholder,
		service_config.SubnetworkAttr, TestSubnetwork,
		service_config.CpuAllocationAttr, TestCpuAllocation,
		service_config.MemoryAllocationAttr, TestMemoryAllocation,*/
	)
}

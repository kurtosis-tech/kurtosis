package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/set_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type setServiceTestCase struct {
	*testing.T
	serviceNetwork               *service_network.MockServiceNetwork
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestSetService() {
	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()
	enclaveDb := getEnclaveDBForTest(suite.T())
	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(suite.T(), err)
	suite.interpretationTimeValueStore = interpretationTimeValueStore

	testServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, nil, nil, []string{}, []string{}, map[string]string{}, nil, nil, 0, 0, "IP-ADDRESS", 0, 0, map[string]string{}, nil, nil, nil, image_download_mode.ImageDownloadMode_Always, true, false, []string{})
	require.NoError(suite.T(), err)
	suite.interpretationTimeValueStore.PutServiceConfig(testServiceName, testServiceConfig)

	suite.run(&setServiceTestCase{
		T:                            suite.T(),
		serviceNetwork:               suite.serviceNetwork,
		packageContentProvider:       suite.packageContentProvider,
		interpretationTimeValueStore: suite.interpretationTimeValueStore,
	})
}

func (t *setServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return set_service.NewSetService(
		t.serviceNetwork,
		t.interpretationTimeValueStore,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		image_download_mode.ImageDownloadMode_Missing)
}

func (t *setServiceTestCase) GetStarlarkCode() string {
	serviceConfig := fmt.Sprintf("ServiceConfig(image=%q)", testContainerImageName)
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, set_service.SetServiceBuiltinName, set_service.ServiceNameArgName, testServiceName, set_service.SetServiceConfigArgName, serviceConfig)
}

func (t *setServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *setServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Set service config on service '%s'.", testServiceName)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}

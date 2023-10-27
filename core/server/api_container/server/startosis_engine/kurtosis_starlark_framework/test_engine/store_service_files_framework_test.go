package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type storeServiceFilesTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func (suite *KurtosisPlanInstructionTestSuite) TestStoreServiceFiles() {
	suite.serviceNetwork.EXPECT().CopyFilesFromService(
		mock.Anything,
		string(testServiceName),
		testSrcPath,
		testArtifactName,
	).Times(1).Return(
		testArtifactUuid,
		nil,
	)

	suite.run(&storeServiceFilesTestCase{
		T:              suite.T(),
		serviceNetwork: suite.serviceNetwork,
	})
}

func (t *storeServiceFilesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return store_service_files.NewStoreServiceFiles(t.serviceNetwork)
}

func (t *storeServiceFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)", store_service_files.StoreServiceFilesBuiltinName, store_service_files.ServiceNameArgName, testServiceName, store_service_files.SrcArgName, testSrcPath, store_service_files.ArtifactNameArgName, testArtifactName)
}

func (t *storeServiceFilesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *storeServiceFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(testArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", testArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

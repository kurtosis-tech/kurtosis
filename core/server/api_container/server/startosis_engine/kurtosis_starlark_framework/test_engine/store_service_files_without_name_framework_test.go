package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type storeServiceFilesWithoutNameTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func (suite *KurtosisPlanInstructionTestSuite) TestStoreServiceFilesWithoutName() {
	suite.serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Times(1).Return(
		mockedFileArtifactName,
		nil,
	)

	suite.serviceNetwork.EXPECT().CopyFilesFromService(
		mock.Anything,
		string(testServiceName),
		testSrcPath,
		mockedFileArtifactName,
	).Times(1).Return(
		testArtifactUuid,
		nil,
	)

	suite.run(&storeServiceFilesWithoutNameTestCase{
		T:              suite.T(),
		serviceNetwork: suite.serviceNetwork,
	})
}

func (t *storeServiceFilesWithoutNameTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return store_service_files.NewStoreServiceFiles(t.serviceNetwork)
}

func (t *storeServiceFilesWithoutNameTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", store_service_files.StoreServiceFilesBuiltinName, store_service_files.ServiceNameArgName, testServiceName, store_service_files.SrcArgName, testSrcPath)
}

func (t *storeServiceFilesWithoutNameTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *storeServiceFilesWithoutNameTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(mockedFileArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", mockedFileArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

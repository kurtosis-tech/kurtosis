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

type storeServiceFilesWithoutNameTestCase struct {
	*testing.T
}

func newStoreServiceFilesWithoutNameTestCase(t *testing.T) *storeServiceFilesWithoutNameTestCase {
	return &storeServiceFilesWithoutNameTestCase{
		T: t,
	}
}

func (t *storeServiceFilesWithoutNameTestCase) GetId() string {
	return store_service_files.StoreServiceFilesBuiltinName
}

func (t *storeServiceFilesWithoutNameTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Times(1).Return(
		mockedFileArtifactName,
		nil,
	)

	serviceNetwork.EXPECT().CopyFilesFromService(
		mock.Anything,
		string(TestServiceName),
		TestSrcPath,
		mockedFileArtifactName,
	).Times(1).Return(
		TestArtifactUuid,
		nil,
	)

	return store_service_files.NewStoreServiceFiles(serviceNetwork)
}

func (t *storeServiceFilesWithoutNameTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", store_service_files.StoreServiceFilesBuiltinName, store_service_files.ServiceNameArgName, TestServiceName, store_service_files.SrcArgName, TestSrcPath)
}

func (t *storeServiceFilesWithoutNameTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(mockedFileArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", mockedFileArtifactName, TestArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

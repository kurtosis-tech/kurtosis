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
}

func newStoreServiceFilesTestCase(t *testing.T) *storeServiceFilesTestCase {
	return &storeServiceFilesTestCase{
		T: t,
	}
}

func (t *storeServiceFilesTestCase) GetId() string {
	return store_service_files.StoreServiceFilesBuiltinName
}

func (t *storeServiceFilesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().CopyFilesFromService(
		mock.Anything,
		string(TestServiceName),
		TestSrcPath,
		TestArtifactName,
	).Times(1).Return(
		TestArtifactUuid,
		nil,
	)

	return store_service_files.NewStoreServiceFiles(serviceNetwork)
}

func (t *storeServiceFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)", store_service_files.StoreServiceFilesBuiltinName, store_service_files.ServiceNameArgName, TestServiceName, store_service_files.SrcArgName, TestSrcPath, store_service_files.ArtifactNameArgName, TestArtifactName)
}

func (t *storeServiceFilesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *storeServiceFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(TestArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", TestArtifactName, TestArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

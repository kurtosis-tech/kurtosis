package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	serviceName  = service.ServiceName("test-service")
	artifactName = "test-artifact"
	src          = "/path/to/file.txt"

	fileArtifactUuid = enclave_data_directory.FilesArtifactUUID("test-artifact-uuid")
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
		string(serviceName),
		src,
		artifactName,
	).Times(1).Return(
		fileArtifactUuid,
		nil,
	)

	return store_service_files.NewStoreServiceFiles(serviceNetwork)
}

func (t *storeServiceFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)", store_service_files.StoreServiceFilesBuiltinName, store_service_files.ServiceNameArgName, serviceName, store_service_files.SrcArgName, src, store_service_files.ArtifactNameArgName, artifactName)
}

func (t *storeServiceFilesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *storeServiceFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(artifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", artifactName, fileArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

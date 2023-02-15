package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	uploadFiles_artifactName     = "artifact-name"
	uploadFiles_src              = "github.com/kurtosistech/my-package/my-file.md"
	uploadFiles_fileArtifactUuid = enclave_data_directory.FilesArtifactUUID("file-artifact-uuid")
)

type uploadFilesTestCase struct {
	*testing.T
}

func newUploadFilesTestCase(t *testing.T) *uploadFilesTestCase {
	return &uploadFilesTestCase{
		T: t,
	}
}

func (t *uploadFilesTestCase) GetId() string {
	return upload_files.UploadFilesBuiltinName
}

func (t *uploadFilesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	require.Nil(t, packageContentProvider.AddFileContent(uploadFiles_src, "Hello World!"))

	serviceNetwork.EXPECT().UploadFilesArtifact(
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		uploadFiles_artifactName,
	).Times(1).Return(
		uploadFiles_fileArtifactUuid,
		nil,
	)

	return upload_files.NewUploadFiles(serviceNetwork, packageContentProvider)
}

func (t uploadFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, uploadFiles_src, upload_files.ArtifactNameArgName, uploadFiles_artifactName)
}

func (t *uploadFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(uploadFiles_artifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", uploadFiles_artifactName, uploadFiles_fileArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

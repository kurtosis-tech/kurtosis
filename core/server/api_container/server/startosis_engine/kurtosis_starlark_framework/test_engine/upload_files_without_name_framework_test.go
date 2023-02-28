package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type uploadFilesWithoutNameTestCase struct {
	*testing.T
}

func newUploadFilesWithoutNameTestCase(t *testing.T) *uploadFilesWithoutNameTestCase {
	return &uploadFilesWithoutNameTestCase{
		T: t,
	}
}

func (t *uploadFilesWithoutNameTestCase) GetId() string {
	return upload_files.UploadFilesBuiltinName
}

func (t *uploadFilesWithoutNameTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	require.Nil(t, packageContentProvider.AddFileContent(uploadFiles_src, "Hello World!"))

	serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Times(1).Return(
		mockedFileArtifactName,
		nil,
	)

	serviceNetwork.EXPECT().UploadFilesArtifact(
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		mockedFileArtifactName,
	).Times(1).Return(
		uploadFiles_fileArtifactUuid,
		nil,
	)

	return upload_files.NewUploadFiles(serviceNetwork, packageContentProvider)
}

func (t uploadFilesWithoutNameTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, uploadFiles_src)
}

func (t *uploadFilesWithoutNameTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(mockedFileArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", mockedFileArtifactName, uploadFiles_fileArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

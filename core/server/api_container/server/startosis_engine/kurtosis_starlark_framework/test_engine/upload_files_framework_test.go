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
	require.Nil(t, packageContentProvider.AddFileContent(TestSrcPath, "Hello World!"))

	serviceNetwork.EXPECT().UploadFilesArtifact(
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		TestArtifactName,
	).Times(1).Return(
		TestArtifactUuid,
		nil,
	)

	return upload_files.NewUploadFiles(serviceNetwork, packageContentProvider)
}

func (t uploadFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, TestSrcPath, upload_files.ArtifactNameArgName, TestArtifactName)
}

func (t *uploadFilesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(TestArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", TestArtifactName, TestArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

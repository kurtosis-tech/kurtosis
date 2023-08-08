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

type uploadFilesUpdateTestCase struct {
	*testing.T
}

func newUploadFilesUpdateTestCase(t *testing.T) *uploadFilesUpdateTestCase {
	return &uploadFilesUpdateTestCase{
		T: t,
	}
}

func (t *uploadFilesUpdateTestCase) GetId() string {
	return upload_files.UploadFilesBuiltinName
}

func (t *uploadFilesUpdateTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	require.Nil(t, packageContentProvider.AddFileContent(TestModuleFileName, "Hello World!"))

	serviceNetwork.EXPECT().GetFilesArtifactMd5(
		TestArtifactName,
	).Times(1).Return(
		TestArtifactUuid,
		[]byte{},
		true,
		nil,
	)
	serviceNetwork.EXPECT().UpdateFilesArtifact(
		TestArtifactUuid,
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		mock.Anything, // and same for the hash
	).Times(1).Return(
		nil,
	)

	return upload_files.NewUploadFiles(serviceNetwork, packageContentProvider)
}

func (t uploadFilesUpdateTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, TestModuleFileName, upload_files.ArtifactNameArgName, TestArtifactName)
}

func (t *uploadFilesUpdateTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesUpdateTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(TestArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' with artifact UUID '%s' updated", TestArtifactName, TestArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

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

type uploadFilesTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFiles() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(testModuleFileName, "Hello World!"))

	suite.serviceNetwork.EXPECT().GetFilesArtifactMd5(
		testArtifactName,
	).Times(1).Return(
		enclave_data_directory.FilesArtifactUUID(""),
		nil,
		false,
		nil,
	)
	suite.serviceNetwork.EXPECT().UploadFilesArtifact(
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		mock.Anything, // and same for the hash
		testArtifactName,
	).Times(1).Return(
		testArtifactUuid,
		nil,
	)

	suite.run(&uploadFilesTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *uploadFilesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return upload_files.NewUploadFiles(testModulePackageId, t.serviceNetwork, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *uploadFilesTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, testModuleRelativeLocator, upload_files.ArtifactNameArgName, testArtifactName)
}

func (t *uploadFilesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(testArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", testArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

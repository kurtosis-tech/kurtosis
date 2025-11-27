package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type uploadFilesUpdateTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFilesUpdate() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(testModuleFileName, "Hello World!"))

	suite.serviceNetwork.EXPECT().GetFilesArtifactMd5(
		testArtifactName,
	).Times(1).Return(
		testArtifactUuid,
		[]byte{},
		true,
		nil,
	)
	suite.serviceNetwork.EXPECT().UpdateFilesArtifact(
		testArtifactUuid,
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		mock.Anything, // and same for the hash
	).Times(1).Return(
		nil,
	)

	suite.run(&uploadFilesUpdateTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *uploadFilesUpdateTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return upload_files.NewUploadFiles(testModulePackageId, t.serviceNetwork, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *uploadFilesUpdateTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, testModuleRelativeLocator, upload_files.ArtifactNameArgName, testArtifactName)
}

func (t *uploadFilesUpdateTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesUpdateTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(testArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' with artifact UUID '%s' updated", testArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

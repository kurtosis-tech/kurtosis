package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/dzobbe/PoTE-kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type uploadFilesWithoutNameTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFilesWithoutName() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(testModuleFileName, "Hello World!"))

	suite.serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Times(1).Return(
		mockedFileArtifactName,
		nil,
	)

	suite.serviceNetwork.EXPECT().GetFilesArtifactMd5(
		mockedFileArtifactName,
	).Times(1).Return(
		enclave_data_directory.FilesArtifactUUID(""),
		nil,
		false,
		nil,
	)
	suite.serviceNetwork.EXPECT().UploadFilesArtifact(
		mock.Anything, // data gets written to disk and compressed to it's a bit tricky to replicate here.
		mock.Anything, // and same for the hash
		mockedFileArtifactName,
	).Times(1).Return(
		testArtifactUuid,
		nil,
	)

	suite.run(&uploadFilesWithoutNameTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *uploadFilesWithoutNameTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return upload_files.NewUploadFiles(testModulePackageId, t.serviceNetwork, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *uploadFilesWithoutNameTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, testModuleRelativeLocator)
}

func (t *uploadFilesWithoutNameTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesWithoutNameTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(mockedFileArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", mockedFileArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

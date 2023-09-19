package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type uploadFilesUpdateTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFilesUpdate() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(TestModuleFileName, "Hello World!"))

	suite.serviceNetwork.EXPECT().GetFilesArtifactMd5(
		TestArtifactName,
	).Times(1).Return(
		TestArtifactUuid,
		[]byte{},
		true,
		nil,
	)
	suite.serviceNetwork.EXPECT().UpdateFilesArtifact(
		TestArtifactUuid,
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
	return upload_files.NewUploadFiles("", t.serviceNetwork, t.packageContentProvider)
}

func (t *uploadFilesUpdateTestCase) GetStarlarkCode() string {
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

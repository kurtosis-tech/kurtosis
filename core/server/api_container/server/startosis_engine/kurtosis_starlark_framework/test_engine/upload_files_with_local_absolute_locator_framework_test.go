package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"testing"
)

const (
	uploadFilesWithLocalAbsoluteLocatorExpectedErrorMsg = "Tried to convert locator 'github.com/kurtosistech/test-package/helpers.star' into absolute locator but failed\n\tCaused by: Cannot use local absolute locators"
)

type uploadFilesWithLocalAbsoluteLocatorTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFilesWithLocalAbsoluteLocatorShouldNotBeValid() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(testModuleFileName, "Hello World!"))

	suite.runShouldFail(
		testModulePackageId,
		&uploadFilesWithLocalAbsoluteLocatorTestCase{
			T:                      suite.T(),
			serviceNetwork:         suite.serviceNetwork,
			packageContentProvider: suite.packageContentProvider,
		},
		uploadFilesWithLocalAbsoluteLocatorExpectedErrorMsg,
	)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return upload_files.NewUploadFiles(testModulePackageId, t.serviceNetwork, t.packageContentProvider, testNoPackageReplaceOptions)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, testModuleFileName, upload_files.ArtifactNameArgName, testArtifactName)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {

}

package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"testing"
)

const (
	uploadFilesWithLocalAbsoluteLocatorExpectedErrorMsg = "Cannot construct 'upload_files' from the provided arguments.\n\tCaused by: The following argument(s) could not be parsed or did not pass validation: {\"src\":\"The locator '\\\"github.com/kurtosistech/test-package/helpers.star\\\"' set in attribute 'src' is not a 'local relative locator'. Local absolute locators are not allowed you should modified it to be a valid 'local relative locator'\"}"
)

type uploadFilesWithLocalAbsoluteLocatorTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider startosis_packages.PackageContentProvider
}

func (suite *KurtosisPlanInstructionTestSuite) TestUploadFilesWithLocalAbsoluteLocatorShouldNotBeValid() {
	suite.Require().Nil(suite.packageContentProvider.AddFileContent(TestModuleFileName, "Hello World!"))

	suite.runShouldFail(
		&uploadFilesWithLocalAbsoluteLocatorTestCase{
			T:                      suite.T(),
			serviceNetwork:         suite.serviceNetwork,
			packageContentProvider: suite.packageContentProvider,
		},
		uploadFilesWithLocalAbsoluteLocatorExpectedErrorMsg,
	)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return upload_files.NewUploadFiles(TestModulePackageId, t.serviceNetwork, t.packageContentProvider, TestNoPackageReplaceOptions)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q)", upload_files.UploadFilesBuiltinName, upload_files.SrcArgName, TestModuleFileName, upload_files.ArtifactNameArgName, TestArtifactName)
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *uploadFilesWithLocalAbsoluteLocatorTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {

}

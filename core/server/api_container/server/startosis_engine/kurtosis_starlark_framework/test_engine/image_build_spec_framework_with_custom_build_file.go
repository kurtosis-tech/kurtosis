package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type imageBuildSpecWithCustomBuildFileTest struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestImageBuildSpecTestWithCustomBuildFile() {
	suite.packageContentProvider.EXPECT().
		GetAbsoluteLocator(testModulePackageId, testModuleMainFileLocator, testBuildContextDir, testNoPackageReplaceOptions).
		Times(1).
		Return(testBuildContextLocator, nil)

	suite.packageContentProvider.EXPECT().
		GetOnDiskAbsolutePackageFilePath(testContainerImageCustomLocator).
		Times(1).
		Return(testOnDiskContainerImageCustomPath, nil)

	suite.run(&imageBuildSpecWithCustomBuildFileTest{
		T:                      suite.T(),
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *imageBuildSpecWithCustomBuildFileTest) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q)",
		service_config.ImageBuildSpecTypeName,
		service_config.BuiltImageNameAttr,
		testContainerImageName,
		service_config.BuildContextAttr,
		testBuildContextDir,
		service_config.BuildFileAttr,
		testCustomBuildFile,
		service_config.TargetStageAttr,
		testEmptyTargetStage)
}

func (t *imageBuildSpecWithCustomBuildFileTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	imageBuildSpecStarlark, ok := typeValue.(*service_config.ImageBuildSpec)
	require.True(t, ok)

	imageBuildSpec, err := imageBuildSpecStarlark.ToKurtosisType(
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, err)
	require.Equal(t, testOnDiskContainerImageCustomPath, imageBuildSpec.GetContainerImageFilePath())
	require.Equal(t, testOnDiskContextDirPath, imageBuildSpec.GetBuildContextDir())
	require.Equal(t, testEmptyTargetStage, imageBuildSpec.GetTargetStage())
}

package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type imageBuildSpecWithBuildArgsTest struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestImageBuildSpecWithBuildArgsTest() {
	suite.packageContentProvider.EXPECT().
		GetAbsoluteLocator(testModulePackageId, testModuleMainFileLocator, testBuildContextDir, testNoPackageReplaceOptions).
		Times(1).
		Return(testBuildContextLocator, nil)

	suite.packageContentProvider.EXPECT().
		GetOnDiskAbsolutePackageFilePath(testContainerImageLocator).
		Times(1).
		Return(testOnDiskContainerImagePath, nil)

	suite.run(&imageBuildSpecWithBuildArgsTest{
		T:                      suite.T(),
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *imageBuildSpecWithBuildArgsTest) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%s)",
		service_config.ImageBuildSpecTypeName,
		service_config.BuiltImageNameAttr,
		testContainerImageName,
		service_config.BuildContextAttr,
		testBuildContextDir,
		service_config.BuildArgsAttr,
		fmt.Sprintf("{%q: %q, %q: %q}", testBuildArgName1, testBuildArgValue1, testBuildArgName2, testBuildArgValue2))
}

func (t *imageBuildSpecWithBuildArgsTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	imageBuildSpecStarlark, ok := typeValue.(*service_config.ImageBuildSpec)
	require.True(t, ok)

	imageBuildSpec, err := imageBuildSpecStarlark.ToKurtosisType(
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, err)
	require.Equal(t, testOnDiskContainerImagePath, imageBuildSpec.GetContainerImageFilePath())
	require.Equal(t, testOnDiskContextDirPath, imageBuildSpec.GetBuildContextDir())

	expectedBuildArgs := map[string]string{
		testBuildArgName1: testBuildArgValue1,
		testBuildArgName2: testBuildArgValue2,
	}
	require.Equal(t, expectedBuildArgs, imageBuildSpec.GetBuildArgs())
}

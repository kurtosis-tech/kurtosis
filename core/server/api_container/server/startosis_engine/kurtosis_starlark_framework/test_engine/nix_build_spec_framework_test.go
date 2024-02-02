package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type nixBuildSpecTest struct {
	*testing.T

	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestNixBuildSpecTest() {
	suite.packageContentProvider.EXPECT().
		GetAbsoluteLocator(testModulePackageId, testModuleMainFileLocator, testBuildContextDir, testNoPackageReplaceOptions).
		Times(1).
		Return(testOnDiskNixContextDirPath, nil)

	suite.packageContentProvider.EXPECT().
		GetOnDiskAbsolutePackageFilePath(testNixFlakeLocator).
		Times(1).
		Return(testOnDiskNixFlakePath, nil)

	suite.run(&nixBuildSpecTest{
		T:                      suite.T(),
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *nixBuildSpecTest) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q)",
		service_config.NixBuildSpecTypeName,
		service_config.FlakeLocationDir,
		testNixFlakeLocationDir,
		service_config.NixContextAttr,
		testNixContextDir,
		service_config.FlakeOutputAttr,
		testNixFlakeOutput)
}

func (t *nixBuildSpecTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	nixBuildSpecStarlark, ok := typeValue.(*service_config.NixBuildSpec)
	require.True(t, ok)

	nixBuildSpec, err := nixBuildSpecStarlark.ToKurtosisType(
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, err)
	require.Equal(t, testOnDiskNixFlakePath, nixBuildSpec.GetNixFlakeFilePath())
	require.Equal(t, testOnDiskNixContextDirPath, nixBuildSpec.GetBuildContextDir())
	require.Equal(t, testNixFlakeOutput, nixBuildSpec.GetFlakeOutput())
}

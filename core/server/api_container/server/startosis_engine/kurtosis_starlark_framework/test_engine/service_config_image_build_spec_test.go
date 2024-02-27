package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type serviceConfigImageBuildSpecTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithImageBuildSpec() {
	suite.packageContentProvider.EXPECT().
		GetAbsoluteLocator(testModulePackageId, testModuleMainFileLocator, testBuildContextDir, testNoPackageReplaceOptions).
		Times(1).
		Return(testBuildContextLocator, nil)

	suite.packageContentProvider.EXPECT().
		GetOnDiskAbsolutePackageFilePath(testContainerImageLocator).
		Times(1).
		Return(testOnDiskContainerImagePath, nil)

	suite.run(&serviceConfigImageBuildSpecTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigImageBuildSpecTestCase) GetStarlarkCode() string {
	imageBuildSpec := fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q)",
		service_config.ImageBuildSpecTypeName,
		service_config.BuiltImageNameAttr,
		testContainerImageName,
		service_config.BuildContextAttr,
		testBuildContextDir,
		service_config.BuildFileAttr,
		testEmptyBuildFile,
		service_config.TargetStageAttr,
		testTargetStage)
	return fmt.Sprintf("%s(%s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, imageBuildSpec)
}

func (t *serviceConfigImageBuildSpecTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, interpretationErr)

	expectedImageBuildSpec := image_build_spec.NewImageBuildSpec(
		testOnDiskContextDirPath,
		testOnDiskContainerImagePath,
		testTargetStage)
	expectedServiceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		expectedImageBuildSpec,
		nil,
		nil,
		map[string]*port_spec.PortSpec{},
		map[string]*port_spec.PortSpec{},
		nil,
		nil,
		map[string]string{},
		nil,
		nil,
		0,
		0,
		service_config.DefaultPrivateIPAddrPlaceholder,
		0,
		0,
		map[string]string{},
		nil,
		nil,
		map[string]string{},
	)
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfig)
	require.Equal(t, expectedImageBuildSpec, serviceConfig.GetImageBuildSpec())
}

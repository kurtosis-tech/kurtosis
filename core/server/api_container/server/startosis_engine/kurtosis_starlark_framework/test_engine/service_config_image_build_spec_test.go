package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"

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
		Return(testModulePackageAbsoluteLocator, nil)

	suite.packageContentProvider.EXPECT().
		GetOnDiskAbsolutePackageFilePath(testContainerImageAbsoluteLocator).
		Times(1).
		Return(testOnDiskContainerImagePath, nil)

	suite.run(&serviceConfigImageBuildSpecTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigImageBuildSpecTestCase) GetStarlarkCode() string {
	imageBuildSpec := fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q, %s=%s)",
		service_config.ImageBuildSpecTypeName,
		service_config.BuiltImageNameAttr,
		testContainerImageName,
		service_config.BuildContextAttr,
		testBuildContextDir,
		service_config.BuildFileAttr,
		defaultBuildFile,
		service_config.TargetStageAttr,
		testTargetStage,
		service_config.BuildArgsAttr,
		fmt.Sprintf("{%q: %q, %q: %q}", testBuildArgName1, testBuildArgValue1, testBuildArgName2, testBuildArgValue2))
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
		testNoPackageReplaceOptions,
		image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, interpretationErr)

	expectedBuildArgs := map[string]string{
		testBuildArgName1: testBuildArgValue1,
		testBuildArgName2: testBuildArgValue2,
	}
	expectedImageBuildSpec := image_build_spec.NewImageBuildSpec(
		testOnDiskContextDirPath,
		testOnDiskContainerImagePath,
		testTargetStage,
		defaultBuildFile,
		expectedBuildArgs)
	expectedServiceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		expectedImageBuildSpec,
		nil,                              // imageRegistrySpec
		nil,                              // nixBuildSpec
		map[string]*port_spec.PortSpec{}, // privatePorts
		map[string]*port_spec.PortSpec{}, // publicPorts
		nil,                              // entrypointArgs
		nil,                              // cmdArgs
		map[string]string{},              // envVars
		nil,                              // filesArtifactExpansion
		nil,                              // persistentDirectories
		0,                                // cpuAllocationMillicpus
		0,                                // memoryAllocationMegabytes
		service_config.DefaultPrivateIPAddrPlaceholder,
		0,                   // minCpuAllocationMilliCpus
		0,                   // minMemoryAllocationMegabytes
		map[string]string{}, // labels
		map[string]string{}, // ingressAnnotations
		nil,                 // ingressClassName
		nil,                 // ingressHost
		nil,                 // ingressTLSHost
		nil,                 // user
		nil,                 // tolerations
		map[string]string{}, // nodeSelectors
		image_download_mode.ImageDownloadMode_Missing,
		true, // waitForPorts
	)
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfig)
	require.Equal(t, expectedImageBuildSpec, serviceConfig.GetImageBuildSpec())
}

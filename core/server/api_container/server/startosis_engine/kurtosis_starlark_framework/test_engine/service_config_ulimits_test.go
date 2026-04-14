package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type serviceConfigUlimitsTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithUlimits() {
	suite.run(&serviceConfigUlimitsTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigUlimitsTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s=%s(%s={%q: %d, %q: %d}))",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.GpuAttr,
		service_config.GpuConfigTypeName,
		service_config.GpuConfigUlimitsAttr,
		"memlock", -1,
		"nofile", 65536)
}

func (t *serviceConfigUlimitsTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfigResult, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, interpretationErr)

	expectedUlimits := map[string]int64{"memlock": -1, "nofile": 65536}
	expectedServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true, false, []string{}, false, service.NewGpuConfig(0, nil, 0, expectedUlimits, service.DefaultDockerGpuDriver, service.DefaultK8sGpuResourceName))
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfigResult)
	require.Equal(t, expectedUlimits, serviceConfigResult.GetGpuConfig().GetUlimits())
}

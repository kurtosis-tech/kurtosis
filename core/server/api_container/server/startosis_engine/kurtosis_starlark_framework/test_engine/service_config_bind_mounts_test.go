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

type serviceConfigBindMountsTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithBindMounts() {
	suite.run(&serviceConfigBindMountsTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigBindMountsTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q, %s={%q: %q})",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.BindMountsAttr, "/var/run/docker.sock", "/var/run/docker.sock")
}

func (t *serviceConfigBindMountsTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
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

	expectedBindMounts := map[string]string{"/var/run/docker.sock": "/var/run/docker.sock"}
	expectedServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true, false, []string{}, false, service.NewGpuConfig(0, nil, 0, nil, "", ""))
	require.NoError(t, err)
	expectedServiceConfig.SetBindMounts(expectedBindMounts)
	require.Equal(t, expectedServiceConfig, serviceConfigResult)
	require.Equal(t, expectedBindMounts, serviceConfigResult.GetBindMounts())
}

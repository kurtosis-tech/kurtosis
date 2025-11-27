package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type serviceConfigMinimalTestCase struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigMinimal() {
	suite.run(&serviceConfigMinimalTestCase{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName)
}

func (t *serviceConfigMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModulePackageId,
		testModuleMainFileLocator,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, interpretationErr)

	expectedServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true, false, []string{})
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfig)
	require.Nil(t, serviceConfig.GetImageBuildSpec())
}

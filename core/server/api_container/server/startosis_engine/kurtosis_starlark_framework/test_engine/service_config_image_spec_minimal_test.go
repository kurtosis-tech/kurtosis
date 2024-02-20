package test_engine

import (
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
)

type serviceConfigImageSpecMinimalTest struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithImageSpecImageOnly() {
	suite.run(&serviceConfigImageSpecMinimalTest{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigImageSpecMinimalTest) GetStarlarkCode() string {
	imageSpec := fmt.Sprintf("%s(%s=%q)",
		service_config.ImageSpecTypeName,
		service_config.ImageAttr,
		testContainerImageName,
	)
	return fmt.Sprintf("%s(%s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, imageSpec)
}

func (t *serviceConfigImageSpecMinimalTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, interpretationErr)

	expectedImageRegistrySpec := image_spec.NewImagSpec(testContainerImageName, "", "", "")
	expectedServiceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		expectedImageRegistrySpec,
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
	require.Equal(t, expectedImageRegistrySpec, serviceConfig.GetImageRegistrySpec())
}

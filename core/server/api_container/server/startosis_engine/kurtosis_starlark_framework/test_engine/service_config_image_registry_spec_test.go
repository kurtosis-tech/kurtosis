package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	"testing"
)

type serviceConfigImageRegistrySpecTest struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithImageRegistrySpec() {
	suite.run(&serviceConfigImageRegistrySpecTest{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigImageRegistrySpecTest) GetStarlarkCode() string {
	imageRegistrySpec := fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q)",
		service_config.ImageRegistrySpecTypeName,
		service_config.ImageAttr,
		testContainerImageName,
		service_config.RegistryAddrAttr,
		testRegistryAddr,
		service_config.RegistryUsernameAttr,
		testRegistryUsername,
		service_config.RegistryPasswordAttr,
		testRegistryPassword,
	)
	return fmt.Sprintf("%s(%s=%s)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, imageRegistrySpec)
}

func (t *serviceConfigImageRegistrySpecTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions)
	require.Nil(t, interpretationErr)

	expectedImageRegistrySpec := image_registry_spec.NewImageRegistrySpec(testContainerImageName, testRegistryUsername, testRegistryPassword, testRegistryAddr)
	expectedServiceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		expectedImageRegistrySpec,
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
	)
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfig)
	require.Equal(t, expectedImageRegistrySpec, serviceConfig.GetImageRegistrySpec())
}

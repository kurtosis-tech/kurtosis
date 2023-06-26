package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/require"
	"testing"
)

type serviceConfigMinimalTestCase struct {
	*testing.T
	serviceNetwork *service_network.MockServiceNetwork
}

func newServiceConfigMinimalTestCase(t *testing.T) *serviceConfigMinimalTestCase {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	return &serviceConfigMinimalTestCase{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t *serviceConfigMinimalTestCase) GetId() string {
	return service_config.ServiceConfigTypeName
}

func (t *serviceConfigMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, TestContainerImageName)
}

func (t *serviceConfigMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType(t.serviceNetwork)
	require.Nil(t, err)

	expectedServiceConfig := service.NewServiceConfig(
		TestContainerImageName,
		map[string]*port_spec.PortSpec{},
		map[string]*port_spec.PortSpec{},
		nil,
		nil,
		map[string]string{},
		nil,
		0,
		0,
		service_config.DefaultPrivateIPAddrPlaceholder,
		0,
		0,
		service_config.DefaultSubnetwork,
	)
	require.Equal(t, expectedServiceConfig, serviceConfig)
}

package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/require"
	"testing"
)

type serviceConfigMinimalTestCase struct {
	*testing.T
}

func newServiceConfigMinimalTestCase(t *testing.T) *serviceConfigMinimalTestCase {
	return &serviceConfigMinimalTestCase{
		T: t,
	}
}

func (t *serviceConfigMinimalTestCase) GetId() string {
	return service_config.ServiceConfigTypeName
}

func (t *serviceConfigMinimalTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return service_config.NewServiceConfigType()
}

func (t *serviceConfigMinimalTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, TestContainerImageName)
}

func (t *serviceConfigMinimalTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, err := serviceConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedServiceConfig := services.NewServiceConfigBuilder(TestContainerImageName)
	require.Equal(t, expectedServiceConfig.Build(), serviceConfig)
}

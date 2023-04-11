package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/update_service_config"
	"github.com/stretchr/testify/require"
	"testing"
)

type updateServiceConfigTestCase struct {
	*testing.T
}

func newUpdateServiceConfigTestCase(t *testing.T) *updateServiceConfigTestCase {
	return &updateServiceConfigTestCase{
		T: t,
	}
}

func (t *updateServiceConfigTestCase) GetId() string {
	return update_service_config.UpdateServiceConfigTypeName
}

func (t *updateServiceConfigTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%q)", update_service_config.UpdateServiceConfigTypeName, update_service_config.SubnetworkAttr, TestSubnetwork)
}

func (t *updateServiceConfigTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	updateServiceConfigStarlark, ok := typeValue.(*update_service_config.UpdateServiceConfig)
	require.True(t, ok)
	updateServiceConfig, err := updateServiceConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedUpdateServiceConfig := binding_constructors.NewUpdateServiceConfig(string(TestSubnetwork))
	require.Equal(t, expectedUpdateServiceConfig, updateServiceConfig)
}

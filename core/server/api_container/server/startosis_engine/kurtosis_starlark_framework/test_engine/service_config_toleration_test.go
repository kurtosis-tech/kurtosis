package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

type serviceConfigTolerationTest struct {
	*testing.T
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *startosis_packages.MockPackageContentProvider
}

func (suite *KurtosisTypeConstructorTestSuite) TestServiceConfigWithTolerationTest() {
	suite.run(&serviceConfigTolerationTest{
		T:                      suite.T(),
		serviceNetwork:         suite.serviceNetwork,
		packageContentProvider: suite.packageContentProvider,
	})
}

func (t *serviceConfigTolerationTest) GetStarlarkCode() string {
	toleration := fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q, %s=%v)",
		service_config.TolerationTypeName,
		service_config.KeyAttr,
		testTolerationKey,
		service_config.OperatorAttr,
		v1.TolerationOpEqual,
		service_config.ValueAttr,
		testTolerationValue,
		service_config.EffectAttr,
		v1.TaintEffectNoSchedule,
		service_config.TolerationSecondsAttr,
		testTolerationSeconds,
	)
	return fmt.Sprintf("%s(%s=%q, %s=[%s])",
		service_config.ServiceConfigTypeName,
		service_config.ImageAttr, testContainerImageName,
		service_config.TolerationsAttr, toleration)
}

func (t *serviceConfigTolerationTest) Assert(typeValue builtin_argument.KurtosisValueType) {
	serviceConfigStarlark, ok := typeValue.(*service_config.ServiceConfig)
	require.True(t, ok)

	serviceConfig, interpretationErr := serviceConfigStarlark.ToKurtosisType(
		t.serviceNetwork,
		testModuleMainFileLocator,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions, image_download_mode.ImageDownloadMode_Missing)
	require.Nil(t, interpretationErr)
	expectedTolerations := []v1.Toleration{{Key: testTolerationKey, Operator: v1.TolerationOpEqual, Value: testTolerationValue, Effect: v1.TaintEffectNoSchedule, TolerationSeconds: &testTolerationSeconds}}
	expectedServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, expectedTolerations, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true)
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig, serviceConfig)
	require.Equal(t, expectedTolerations, serviceConfig.GetTolerations())
}

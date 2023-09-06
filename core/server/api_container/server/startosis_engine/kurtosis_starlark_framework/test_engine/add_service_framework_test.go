package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_value_serde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

type addServiceTestCase struct {
	*testing.T
}

func newAddServiceTestCase(t *testing.T) *addServiceTestCase {
	return &addServiceTestCase{
		T: t,
	}
}

func (t *addServiceTestCase) GetId() string {
	return add_service.AddServiceBuiltinName
}

func (t *addServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	enclaveDb := getEnclaveDBForTest(t.T)
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(enclaveDb, starlark_value_serde.GetStarlarkValueSerdeForTest())
	require.NoError(t, err)

	serviceNetwork.EXPECT().ExistServiceRegistration(TestServiceName).Times(1).Return(false, nil)
	serviceNetwork.EXPECT().AddService(
		mock.Anything,
		TestServiceName,
		mock.MatchedBy(func(serviceConfig *service.ServiceConfig) bool {
			expectedServiceConfig := service.NewServiceConfig(
				TestContainerImageName,
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
			)

			actualServiceConfig := serviceConfig
			assert.Equal(t, expectedServiceConfig, actualServiceConfig)
			return true
		}),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(TestServiceName, TestServiceUuid, TestEnclaveUuid, nil, string(TestServiceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		nil,
	)

	return add_service.NewAddService(serviceNetwork, runtimeValueStore)
}

func (t *addServiceTestCase) GetStarlarkCode() string {
	serviceConfig := fmt.Sprintf("ServiceConfig(image=%q)", TestContainerImageName)
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, add_service.AddServiceBuiltinName, add_service.ServiceNameArgName, TestServiceName, add_service.ServiceConfigArgName, serviceConfig)
}

func (t *addServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	serviceObj, ok := interpretationResult.(*kurtosis_types.Service)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.NotNil(t, serviceObj)
	expectedServiceObj := fmt.Sprintf(`Service\(name="%v", hostname="{{kurtosis:[0-9a-f]{32}:hostname.runtime_value}}", ip_address="{{kurtosis:[0-9a-f]{32}:ip_address.runtime_value}}", ports={}\)`, TestServiceName)
	require.Regexp(t, expectedServiceObj, serviceObj.String())

	expectedExecutionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", TestServiceName, TestServiceUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}

package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StartosisIntepreterPlanYamlTestSuite struct {
	suite.Suite
	serviceNetwork               *service_network.MockServiceNetwork
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	interpreter *StartosisInterpreter
}

func (suite *StartosisIntepreterPlanYamlTestSuite) SetupTest() {
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	enclaveDb := getEnclaveDBForTest(suite.T())

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(suite.T(), err)
	suite.runtimeValueStore = runtimeValueStore
	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(suite.T(), err)
	suite.interpretationTimeValueStore = interpretationTimeValueStore
	require.NotNil(suite.T(), interpretationTimeValueStore)

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)

	service.NewServiceRegistration(
		testServiceName,
		service.ServiceUUID(fmt.Sprintf("%s-%s", testServiceName, serviceUuidSuffix)),
		mockEnclaveUuid,
		testServiceIpAddress,
		string(testServiceName),
	)
	suite.serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Maybe().Return(mockFileArtifactName, nil)
	suite.serviceNetwork.EXPECT().GetEnclaveUuid().Maybe().Return(enclave.EnclaveUUID(mockEnclaveUuid))
	suite.serviceNetwork.EXPECT().ExistServiceRegistration(testServiceName).Maybe().Return(true, nil)
}

//func TestRunStartosisIntepreterPlanYamlTestSuite(t *testing.T) {
//	suite.Run(t, new(StartosisIntepreterPlanYamlTestSuite))
//}

func (suite *StartosisIntepreterPlanYamlTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestAddService() {
	script := `
def run(plan):
	service_name = "serviceA"
	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = 1234, transport_protocol = "TCP", application_protocol = "http")
		},
	)
	datastore_service = plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	emptyPlanYaml := plan_yaml.CreateEmptyPlan(startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript)
	planYaml, err := instructionsPlan.GenerateYaml(emptyPlanYaml)
	require.NoError(suite.T(), err)

	expectedYaml :=
		`packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- name: serviceA
  image: ` + testContainerImageName + ` 
  ports:
  - name: grpc
    number: 1234
    transportProtocol: TCP
    applicationProtocol: http
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

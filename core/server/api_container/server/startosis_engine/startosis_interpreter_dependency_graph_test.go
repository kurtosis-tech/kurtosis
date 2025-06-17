package startosis_engine

import (
	"context"
	"fmt"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type StartosisIntepreterDependencyGraphTestSuite struct {
	suite.Suite
	serviceNetwork               *service_network.MockServiceNetwork
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	interpreter *StartosisInterpreter
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) SetupTest() {
	// mock package content provider
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	enclaveDb := getEnclaveDBForTest(suite.T())

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	// mock runtime value store
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(suite.T(), err)
	suite.runtimeValueStore = runtimeValueStore

	// mock interpretation time value store
	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(suite.T(), err)
	suite.interpretationTimeValueStore = interpretationTimeValueStore

	// mock service network
	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())
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
	// apiContainerInfo := service_network.NewApiContainerInfo(
	// 	net.IP{},
	// 	mockApicPortNum,
	// 	mockApicVersion)
	// suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)
}

func TestRunStartosisIntepreterDependencyGraphTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisIntepreterDependencyGraphTestSuite))
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *StartosisIntepreterPlanYamlGeneratorTestSuite) TestAddSingleServiceToDependencyGraph() {
	script := `def run(plan):
	config = ServiceConfig(
		image = "ubuntu",
	)
	service = plan.add_service(name = "serviceA", config = config)
`

	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
	}

	inputArgs := `{}`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		inputArgs,
		defaultNonBlockingMode,
		emptyEnclaveComponents,
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

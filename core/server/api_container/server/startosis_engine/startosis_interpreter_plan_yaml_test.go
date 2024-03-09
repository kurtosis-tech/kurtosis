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
	"net"
	"testing"
)

const (
	mockApicPortNum = 1234
	mockApicVersion = "1234"
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
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)
}

func TestRunStartosisIntepreterPlanYamlTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisIntepreterPlanYamlTestSuite))
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestAddServiceWithFilesArtifact() {
	script := `def run(plan, hi_files_artifact):
	service_name = "serviceA"
	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		cmd = ["echo", "Hi"],
		entrypoint = ["sudo", "something"],
		env_vars = {
			"USERNAME": "KURTOSIS"
		},
		ports = {
			"grpc": PortSpec(number = 1234, transport_protocol = "TCP", application_protocol = "http")
		},
		files = {
			"hi.txt": hi_files_artifact,
		},
	)
	datastore_service = plan.add_service(name = service_name, config = config)
`
	inputArgs := `{"hi_files_artifact": "hi-file"}`
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
		emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml :=
		`packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- uuid: "1"
  name: serviceA
  image:
    name: ` + testContainerImageName + `
  command:
  - echo
  - Hi
  entrypoint:
  - sudo
  - something
  envVars:
  - key: USERNAME
    value: KURTOSIS
  ports:
  - name: grpc
    number: 1234
    transportProtocol: TCP
    applicationProtocol: http
  files:
  - mountPath: hi.txt
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestRunShWithFilesArtifacts() {
	script := `def run(plan, hi_files_artifact):
    plan.run_sh(
        run="echo bye > /bye.txt",
        env_vars = {
            "HELLO": "Hello!"
        },
        files = {
            "/hi.txt": hi_files_artifact,
        },
        store=[
            StoreSpec(src="/bye.txt", name="bye-file")
        ]
    )
`
	inputArgs := `{"hi_files_artifact": "hi-file"}`
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
		emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml :=
		`packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
tasks:
- uuid: "1"
  run: echo bye > /bye.txt
  image: bq/jcurl
  envVars:
  - key: HELLO
    value: Hello!
  files:
  - mountPath: /hi.txt
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
- uuid: "3"
  name: bye-file
  files:
  - bye.txt
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

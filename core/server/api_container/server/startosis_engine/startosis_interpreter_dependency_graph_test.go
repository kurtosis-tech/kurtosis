package startosis_engine

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
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
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
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
	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)
}

func TestRunStartosisIntepreterDependencyGraphTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisIntepreterDependencyGraphTestSuite))
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddSingleServiceToDependencyGraph() {
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnFilesArtifactFromRenderTemplate() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):


	artifact_a = plan.render_templates(
		name="artifact_a", 
		config={
			"hi.txt": struct(
				template="{{ .fileA }}",
				data = {
					"fileA": "hi",
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"fileA": artifact_a,
		}
	)
	service = plan.add_service(name = "serviceA", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnAddServiceTemplate() {
	script := `def run(plan):

	config = ServiceConfig(
		image = "ubuntu",
	)
	service_a = plan.add_service(name = "serviceA", config = config)

	config = ServiceConfig(
		image = "ubuntu",
		cmd = [
			"echo {0}".format(service_a.ip_address),
		],
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplate() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .HelloMessage }}",
				data = {
					"HelloMessage": "hi",
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
	service_a = plan.add_service(name = "serviceA", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnAddService() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .ServiceAIpAddress }}",
				data = {
					"ServiceAIpAddress": service_a.ip_address,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnTwoAddServices() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))
	service_b = plan.add_service(name = "serviceB", config = ServiceConfig(image = "ubuntu"))

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .ServiceAIpAddress }} {{ .ServiceBIpAddress }}",
				data = {
					"ServiceAIpAddress": service_a.ip_address,
					"ServiceBIpAddress": service_b.ip_address,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
	service_c = plan.add_service(name = "serviceC", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("1"),
			instructions_plan.ScheduledInstructionUuid("2"),
		},
		instructions_plan.ScheduledInstructionUuid("4"): {
			instructions_plan.ScheduledInstructionUuid("3"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnRunSh() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	task_a_result = plan.run_sh(name = "taskA", run = "echo 'hi'")

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .TaskAResultOutput }} {{ .TaskAResultCode }}",
				data = {
					"TaskAResultOutput": task_a_result.output,
					"TaskAResultCode": task_a_result.code,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnRunShWithFilesArtifact() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	task_a_result = plan.run_sh(name = "taskA", run = "echo 'hi'", store = [StoreSpec(name = "taskAFile", src = "/root/hi.txt")])

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .TaskAResultOutput }} {{ .TaskAResultCode }}",
				data = {
					"TaskAResultOutput": task_a_result.output,
					"TaskAResultCode": task_a_result.code,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
			"/root/another_file.txt": task_a_result.files_artifacts[0],
		}
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestRunShDependsOnService() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .Hi }}",
				data = {
					"Hi": "hi",
				},
			)
		}
	)

	task_a_result = plan.run_sh(name = "taskA", run = "echo Hi")
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnRunPython() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	task_a_result = plan.run_python(name = "taskA", run = "print('hi')")

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .TaskAResultOutput }} {{ .TaskAResultCode }}",
				data = {
					"TaskAResultOutput": task_a_result.output,
					"TaskAResultCode": task_a_result.code,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestAddServiceDependsOnRenderedTemplateDependsOnRunPythonWithFilesArtifact() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	task_a_result = plan.run_python(name = "taskA", run = "print('hi')", store = [StoreSpec(name = "taskAFile", src = "/root/hi.txt")])

	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .TaskAResultOutput }} {{ .TaskAResultCode }}",
				data = {
					"TaskAResultOutput": task_a_result.output,
					"TaskAResultCode": task_a_result.code,
				},
			)
		}
	)

	config = ServiceConfig(
		image = "ubuntu",
		files = {
			"/root/hi.txt": artifact_a,
			"/root/another_file.txt": task_a_result.files_artifacts[0],
		}
	)
	service_b = plan.add_service(name = "serviceB", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestRunPythonDependsOnService() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .Hi }}",
				data = {
					"Hi": "hi",
				},
			)
		}
	)

	task_a_result = plan.run_python(name = "taskA", run = "asfa",files = { "/root/artifact.txt": artifact_a })
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestRunPythonWithArgsDependsOnService() {
	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))

	task_a_result = plan.run_python(
		name = "taskA", 
		run = "import sys; print(sys.argv[1])", 
		args = [service_a.ip_address]
	)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestRunPythonWithPackagesDependsOnFilesArtifact() {
	apiContainerInfo := service_network.NewApiContainerInfo(
		net.IP{},
		mockApicPortNum,
		mockApicVersion)
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	script := `def run(plan):
	artifact_a = plan.render_templates(
		name="another-artifact", 
		config={
			"hi.txt": struct(
				template="{{ .Hi }}",
				data = {
					"Hi": "hi",
				},
			)
		}
	)

	task_a_result = plan.run_python(
		name = "taskA", 
		run = "import requests; print('success')", 
		packages = ["requests"],
		files = {
			"/root/hi.txt": artifact_a,
		}
	)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestRunPythonWithCustomImageDependsOnService() {
	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))

	task_a_result = plan.run_python(
		name = "taskA", 
		run = "print('hi')", 
		image = "python:3.12-alpine"
	)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestExecDependsOnService() {
	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))

	exec_recipe = ExecRecipe(
		command = ["echo", "Hello, world"],
	)
	result = plan.exec(
		service_name = "serviceA",
		recipe = exec_recipe,
	)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestExecOnServiceBDependsOnServiceA() {
	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))
	service_b = plan.add_service(name = "serviceB", config = ServiceConfig(image = "ubuntu"))

	exec_recipe = ExecRecipe(
		command = ["echo", service_a.ip_address],
	)
	result = plan.exec(
		service_name = "serviceB",
		recipe = exec_recipe,
	)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
			instructions_plan.ScheduledInstructionUuid("1"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func (suite *StartosisIntepreterDependencyGraphTestSuite) TestServiceCDependsOnExecCode() {
	script := `def run(plan):
	service_a = plan.add_service(name = "serviceA", config = ServiceConfig(image = "ubuntu"))

	exec_recipe = ExecRecipe(
		command = ["echo", "Hello, world"],
	)
	exec_result = plan.exec(
		service_name = "serviceA",
		recipe = exec_recipe,
	)

	config = ServiceConfig(
		image = "ubuntu",
		cmd = [
			"echo {0}".format(exec_result["code"]),
		],
	)
	service_c = plan.add_service(name = "serviceC", config = config)
`
	expectedDependencyGraph := map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid{
		instructions_plan.ScheduledInstructionUuid("1"): {},
		instructions_plan.ScheduledInstructionUuid("2"): {
			instructions_plan.ScheduledInstructionUuid("1"),
		},
		instructions_plan.ScheduledInstructionUuid("3"): {
			instructions_plan.ScheduledInstructionUuid("2"),
		},
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

	instructionsDependencyGraph := instructionsPlan.GenerateInstructionsDependencyGraph()

	outputDependencyGraphVisual(instructionsDependencyGraph)
	require.Equal(suite.T(), expectedDependencyGraph, instructionsDependencyGraph)
}

func outputDependencyGraphVisual(dependencyGraph map[instructions_plan.ScheduledInstructionUuid][]instructions_plan.ScheduledInstructionUuid) {
	g := simple.NewDirectedGraph()

	nodes := make(map[string]int64)

	for to, fromList := range dependencyGraph {
		if _, ok := nodes[string(to)]; !ok {
			nextID, err := strconv.ParseInt(string(to), 10, 64)
			if err != nil {
				panic(err)
			}
			nodes[string(to)] = nextID
			g.AddNode(simple.Node(nextID))
		}
		for _, from := range fromList {
			if _, ok := nodes[string(from)]; !ok {
				nextID, err := strconv.ParseInt(string(from), 10, 64)
				if err != nil {
					panic(err)
				}
				nodes[string(from)] = nextID
				g.AddNode(simple.Node(nextID))
			}
			g.SetEdge(g.NewEdge(simple.Node(nodes[string(to)]), simple.Node(nodes[string(from)])))
		}
	}

	b, err := dot.Marshal(g, "InstructionsDependencyGraph", "", "  ")
	if err != nil {
		panic(err)
	}

	// Write to file
	if err := os.WriteFile("/Users/tewodrosmitiku/craft/graphs/dependency.dot", b, 0644); err != nil {
		panic(err)
	}

	cmd := exec.Command("dot", "-Tpng", "/Users/tewodrosmitiku/craft/graphs/dependency.dot", "-o", "/Users/tewodrosmitiku/craft/graphs/graph.png")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

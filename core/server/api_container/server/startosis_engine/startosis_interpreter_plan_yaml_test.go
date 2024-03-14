package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
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
            "/root": hi_files_artifact,
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
filesArtifacts:
- uuid: "2"
  name: hi-file
- uuid: "3"
  name: bye-file
  files:
  - /bye.txt
tasks:
- uuid: "1"
  taskType: sh
  command:
  - echo bye > /bye.txt
  image: badouralix/curl-jq
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "2"
      name: hi-file
  store:
  - uuid: "3"
    name: bye-file
  envVar:
  - key: HELLO
    value: Hello!
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestRunPython() {
	script := `def run(plan, hi_files_artifact):
     plan.run_python(
        run = """
    import requests
    response = requests.get("docs.kurtosis.com")
    print(response.status_code)      
    """,
        args = [
           "something" 
        ],
        packages = [
            "selenium",
            "requests",
        ],
        files = {
            "/hi.txt": hi_files_artifact,
        },
        store = [
            StoreSpec(src = "bye.txt", name = "bye-file"),
        ],
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
filesArtifacts:
- uuid: "2"
  name: hi-file
- uuid: "3"
  name: bye-file
  files:
  - bye.txt
tasks:
- uuid: "1"
  taskType: python
  command:
  - "\n    import requests\n    response = requests.get(\"docs.kurtosis.com\")\n    print(response.status_code)
    \     \n    "
  image: python:3.11-alpine
  files:
  - mountPath: /hi.txt
    filesArtifacts:
    - uuid: "2"
      name: hi-file
  store:
  - uuid: "3"
    name: bye-file
  pythonPackages:
  - selenium
  - requests
  pythonArgs:
  - something
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestExec() {
	script := `def run(plan, hi_files_artifact):
	plan.add_service(
		name="db",
		config=ServiceConfig(
			image="postgres:latest",
			env_vars={
				"POSTGRES_DB": "kurtosis",
				"POSTGRES_USER": "kurtosis",
			}, 
			files = {
				"/root": hi_files_artifact,
			}
		)
	)
	result = plan.exec(
		service_name="db",
		recipe=ExecRecipe(command=["echo", "Hello, world"]),
		acceptable_codes=[0],
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- uuid: "1"
  name: db
  image:
    name: postgres:latest
  envVars:
  - key: POSTGRES_DB
    ***REMOVED***
  - key: POSTGRES_USER
    ***REMOVED***
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
tasks:
- uuid: "3"
  taskType: exec
  command:
  - echo
  - Hello, world
  serviceName: db
  acceptableCodes:
  - 0
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestRenderTemplate() {
	script := `def run(plan, args):
    bye_files_artifact = plan.render_templates(
        name="bye-file",
        config={
            "bye.txt": struct(
                template="Bye bye!",
                data={}
            ),
			"fairwell.txt": struct (
				template = "Fair well!",
				data = {}
			),
        }
    )

    plan.run_sh(
        run="cat /root/bye.txt",
        files = {
            "/root": bye_files_artifact,
        },
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
filesArtifacts:
- uuid: "1"
  name: bye-file
  files:
  - bye.txt
  - fairwell.txt
tasks:
- uuid: "2"
  taskType: sh
  command:
  - cat /root/bye.txt
  image: badouralix/curl-jq
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "1"
      name: bye-file
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestAddServiceWithImageBuildSpec() {
	dockerfileModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server/Dockerfile"
	serverModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server"
	dockerfileContents := `RUN ["something"]`
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(dockerfileModulePath, dockerfileContents))
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(serverModulePath, ""))
	packageId := "github.com/kurtosis-tech/plan-yaml-prac"

	script := `def run(plan, hi_files_artifact):
    plan.add_service(
		name="db",
		config=ServiceConfig(
			image = ImageBuildSpec(
				image_name="` + testContainerImageName + `",
				build_context_dir="./server",
				target_stage="builder",
			),
			files = {
				"/root": hi_files_artifact,
			}
		)
	)
`
	inputArgs := `{"hi_files_artifact": "hi-file"}`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(
		context.Background(),
		packageId,
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

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(packageId))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: ` + packageId + `
services:
- uuid: "1"
  name: db
  image:
    name: kurtosistech/example-datastore-server
    buildContextLocator: ./server
    targetStage: builder
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestAddServiceWithImageSpec() {
	dockerfileModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server/Dockerfile"
	serverModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server"
	dockerfileContents := `RUN ["something"]`
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(dockerfileModulePath, dockerfileContents))
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(serverModulePath, ""))
	packageId := "github.com/kurtosis-tech/plan-yaml-prac"

	script := `def run(plan, hi_files_artifact):
    plan.add_service(
		name="db",
		config=ServiceConfig(
			image = ImageSpec(
				image="` + testContainerImageName + `",
				registry = "http://my.registry.io/",
			),
			files = {
				"/root": hi_files_artifact,
			}
		)
	)
`
	inputArgs := `{"hi_files_artifact": "hi-file"}`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(
		context.Background(),
		packageId,
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

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(packageId))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: ` + packageId + `
services:
- uuid: "1"
  name: db
  image:
    name: kurtosistech/example-datastore-server
    registry: http://my.registry.io/
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestUploadFiles() {
	dockerfileModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server/Dockerfile"
	serverModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server"
	dockerfileContents := `RUN ["something"]`
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(dockerfileModulePath, dockerfileContents))
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(serverModulePath, ""))

	packageId := "github.com/kurtosis-tech/plan-yaml-prac"

	script := `def run(plan, args):
    dockerfile_artifact = plan.upload_files(src="./server/Dockerfile", name="dockerfile")

    plan.run_sh(
        run="cat /root/Dockerfile",
        files = {
            "/root": dockerfile_artifact,
        },
    )
`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(
		context.Background(),
		packageId,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		startosis_constants.EmptyInputArgs,
		defaultNonBlockingMode,
		emptyEnclaveComponents,
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(packageId))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: ` + packageId + `
filesArtifacts:
- uuid: "1"
  name: dockerfile
  files:
  - ./server/Dockerfile
tasks:
- uuid: "2"
  taskType: sh
  command:
  - cat /root/Dockerfile
  image: badouralix/curl-jq
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "1"
      name: dockerfile
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestStoreServiceFiles() {
	script := `def run(plan, hi_files_artifact):
    plan.add_service(
        name="db",
        config=ServiceConfig(
            image="postgres:latest",
            cmd=["touch", "bye.txt"],
			files = {
				"/root": hi_files_artifact
			}
        ),
    )

    bye_files_artifact = plan.store_service_files(
        name="bye-file",
        src="bye.txt",
        service_name="db",
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- uuid: "1"
  name: db
  image:
    name: postgres:latest
  command:
  - touch
  - bye.txt
  files:
  - mountPath: /root
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

func (suite *StartosisIntepreterPlanYamlTestSuite) TestRemoveService() {
	script := `def run(plan, hi_files_artifact):
	plan.add_service(
		name="db",
		config=ServiceConfig(
			image="postgres:latest",
			env_vars={
				"POSTGRES_DB": "tedi",
				"POSTGRES_USER": "tedi",
			},
			files = {
				"/root": hi_files_artifact,
			}
		)
	)
	plan.remove_service(name="db")
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
filesArtifacts:
- uuid: "2"
  name: hi-file
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

func (suite *StartosisIntepreterPlanYamlTestSuite) TestFutureReferencesAreSwapped() {
	script := `def run(plan, hi_files_artifact):
	service = plan.add_service(
		name="db",
		config=ServiceConfig(
			image="postgres:latest",
			env_vars={
				"POSTGRES_DB": "kurtosis",
				"POSTGRES_USER": "kurtosis",
			},
			files = {
				"/root": hi_files_artifact,
			}
		)
	)
	execResult = plan.exec(
		service_name="db",
		recipe=ExecRecipe(
			command=["echo", service.ip_address + " " + service.hostname]
		),
		acceptable_codes=[0],
	)	
	runShResult = plan.run_sh(
		run="echo " + execResult["code"] + " " + execResult["output"],
	)
	plan.run_sh(
		run="echo " + runShResult.code + " " + runShResult.output,
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
		emptyInstructionsPlanMask,
		image_download_mode.ImageDownloadMode_Always)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 4, instructionsPlan.Size())

	planYaml, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	require.NoError(suite.T(), err)

	expectedYaml := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- uuid: "1"
  name: db
  image:
    name: postgres:latest
  envVars:
  - key: POSTGRES_DB
    ***REMOVED***
  - key: POSTGRES_USER
    ***REMOVED***
  files:
  - mountPath: /root
    filesArtifacts:
    - uuid: "2"
      name: hi-file
filesArtifacts:
- uuid: "2"
  name: hi-file
tasks:
- uuid: "3"
  taskType: exec
  command:
  - echo
  - '{{ kurtosis.1.ip_address }} {{ kurtosis.1.hostname }}'
  serviceName: db
  acceptableCodes:
  - 0
- uuid: "4"
  taskType: sh
  command:
  - echo {{ kurtosis.3.code }} {{ kurtosis.3.output }}
  image: badouralix/curl-jq
- uuid: "5"
  taskType: sh
  command:
  - echo {{ kurtosis.4.code }} {{ kurtosis.4.output }}
  image: badouralix/curl-jq
`
	require.Equal(suite.T(), expectedYaml, planYaml)
}

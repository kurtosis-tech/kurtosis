package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/file_artifacts_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"os"
	"testing"
)

type PlanYamlGeneratorTestSuite struct {
	suite.Suite
	serviceNetwork               *service_network.MockServiceNetwork
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	kurtosisBackend              *backend_interface.KurtosisBackend
	filesArtifactStore           *enclave_data_directory.FilesArtifactStore
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	interpreter *StartosisInterpreter
	validator   *StartosisValidator
	executor    *StartosisExecutor
	runner      *StartosisRunner
}

func (suite *PlanYamlGeneratorTestSuite) SetupTest() {
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
		51243,
		"134123")
	suite.serviceNetwork.EXPECT().GetApiContainerInfo().Return(apiContainerInfo)

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)

	// mock kurtosis backend?

	// mock files artifact for testing
	timesCalled := 0
	mockedGenerateNameMethod := func() string {
		timesCalled = timesCalled + 1
		if timesCalled == 4 {
			return "last-noun"
		}
		return "adjective-noun"
	}

	artifactIdToUUIDMap := map[string]string{}

	fileArtifactDb, err := file_artifacts_db.GetFileArtifactsDbForTesting(enclaveDb, artifactIdToUUIDMap)
	require.Nil(suite.T(), err)

	suite.filesArtifactStore = enclave_data_directory.NewFilesArtifactStoreForTesting("/", "/", fileArtifactDb, 3, mockedGenerateNameMethod)
	suite.validator = NewStartosisValidator(suite.kurtosisBackend, suite.serviceNetwork, suite.filesArtifactStore)

	// mock the starlark value serde?
	suite.executor = NewStartosisExecutor(nil, suite.runtimeValueStore, enclave_plan_persistence.NewEnclavePlan(), enclaveDb)

	suite.runner = NewStartosisRunner(suite.interpreter, suite.validator, suite.executor)
}

func TestRunPlanYamlGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(PlanYamlGeneratorTestSuite))
}

func (suite *PlanYamlGeneratorTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *PlanYamlGeneratorTestSuite) TestCurrentlyBeingWorkedOn() {
	dockerfileModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server/Dockerfile"
	serverModulePath := "github.com/kurtosis-tech/plan-yaml-prac/server"
	dockerfileContents := `# Use an existing docker image as a base
FROM alpine:latest

# Run commands to install necessary dependencies
RUN apk add --update nodejs npm

# Set the working directory inside the container
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Install app dependencies
RUN npm install

# Expose a port the app runs on
EXPOSE 3000

# Define environment variable
ENV NODE_ENV=production

# Command to run the application
CMD ["node", "app.js"]
`
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(dockerfileModulePath, dockerfileContents))
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(serverModulePath, ""))

	packageId := "github.com/kurtosis-tech/plan-yaml-prac"
	mainFunctionName := ""
	relativePathToMainFile := "main.star"

	serializedScript := `def run(plan, args):
	result = plan.run_sh(
		run="echo some stuff",
	)

	database = plan.add_service(name="database", config=ServiceConfig(
		image="postgres:latest",
		env_vars={
			"VAR_1": result.output,
			"VAR_2": result.code
		}
	))
`
	serializedJsonParams := "{}"
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), packageId, mainFunctionName, noPackageReplaceOptions, relativePathToMainFile, serializedScript, serializedJsonParams, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		packageId,
		suite.packageContentProvider,
		noPackageReplaceOptions,
	)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	err = os.WriteFile("./plan.yml", yamlBytes, 0644)
	require.NoError(suite.T(), err)
}

func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorVerySimpleScript() {
	script := `
def run(plan):
	service_name = "partyService"
	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP", application_protocol = "http")
		},
	)
	datastore_service = plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		suite.packageContentProvider,
		noPackageReplaceOptions)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString :=
		`packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- name: partyService
  image: kurtosistech/example-datastore-server
  ports:
  - name: grpc
    number: 1323
    transportProtocol: TCP
    applicationProtocol: http
`
	require.Equal(suite.T(), expectedYamlString, string(yamlBytes))
}

func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorSimplerScripButNotSoSimple() {
	script := `
def run(plan):
	
	service_name = "partyService"

	config = ServiceConfig(
		image = "` + testContainerImageName + `",
        cmd=["echo", "Hello"],
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP", application_protocol = "http")
		},
		env_vars = {
			"POSTGRES_DB": "tedi",
            "POSTGRES_USERNAME": "dag",
		},
		files = {
			"/usr/": "", # TODO: how do you model a files artifact
 			"/bin/": "", # TODO: how do you model a files artifact
		}
	)
	datastore_service = plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		suite.packageContentProvider,
		noPackageReplaceOptions)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString :=
		`packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- name: partyService
  image: kurtosistech/example-datastore-server
  ports:
  - name: grpc
    number: 1323
    transportProtocol: TCP
    applicationProtocol: http
`
	require.Equal(suite.T(), expectedYamlString, string(yamlBytes))
}

func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorSimpleScript() {
	script := `
service_name = "example-datastore-server"
ports = [1323, 1324, 1325]	

def deploy_datastore_services(plan):
	for i in range(len(ports)):
		unique_service_name = service_name + "-" + str(i)
		plan.print("Adding service " + unique_service_name)
		config = ServiceConfig(
			image = "` + testContainerImageName + `",
			ports = {
				"grpc": PortSpec(
					number = ports[i],
					transport_protocol = "TCP"
				)
			}
		)

		plan.add_service(name = unique_service_name, config = config)

def run(plan):
	plan.print("Starting Startosis script!")
	deploy_datastore_services(plan)
	plan.print("Done!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 8, instructionsPlan.Size())

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		suite.packageContentProvider,
		noPackageReplaceOptions)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- name: partyService
  image: kurtosistech/example-datastore-server
  ports:
  - name: grpc
    number: 1323
    transportProtocol: TCP
    applicationProtocol: http
`
	require.Equal(suite.T(), expectedYamlString, string(yamlBytes))
}

func (suite *PlanYamlGeneratorTestSuite) TestConvertPlanYamlToYamlBytes(t *testing.T) {
	PackageId := "github.com/kurtosis-tech/postgres-package"

	services := []*Service{
		{
			Name: "tedi",
			Uuid: "1",
			Image: &ImageSpec{
				ImageName:           "postgres",
				BuildContextLocator: "",
				TargetStage:         "",
				Registry:            "",
			},
			EnvVars: []*EnvironmentVariable{
				{
					Key:   "kevin",
					Value: "dag",
				},
			},
		},
		{
			Name: "kaleb",
			Uuid: "1",
			Image: &ImageSpec{
				ImageName:           "something",
				BuildContextLocator: "",
				TargetStage:         "",
				Registry:            "",
			},
			EnvVars: []*EnvironmentVariable{
				{
					Key:   "kevin",
					Value: "dag",
				},
			},
		},
	}
	filesArtifacts := []*FilesArtifact{
		{
			Uuid:  "3",
			Name:  "something",
			Files: nil,
		},
	}
	tasks := []*Task{
		{
			TaskType: PYTHON,
			Image:    "jqcurl",
			EnvVars:  []*EnvironmentVariable{},
		},
	}

	planYaml := PlanYaml{
		PackageId:      PackageId,
		Services:       services,
		FilesArtifacts: filesArtifacts,
		Tasks:          tasks,
	}

	yamlBytes, err := convertPlanYamlToYaml(&planYaml)
	require.NoError(t, err)

	expectedYamlString := `packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
- name: partyService
  image: kurtosistech/example-datastore-server
  ports:
  - name: grpc
    number: 1323
    transportProtocol: TCP
    applicationProtocol: http
`
	require.Equal(t, expectedYamlString, string(yamlBytes))
}

func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorPostgresPackageSimplified() {
	packageId := "github.com/kurtosis-tech/postgres-package"
	mainFunctionName := ""
	relativePathToMainFile := "main.star"

	serializedPostgresPackageStarlark := `PORT_NAME = "postgresql"
APPLICATION_PROTOCOL = "postgresql"
PG_DRIVER = "pgsql"

CONFIG_FILE_MOUNT_DIRPATH = "/config"
SEED_FILE_MOUNT_PATH = "/docker-entrypoint-initdb.d"
DATA_DIRECTORY_PATH = "/data/"

CONFIG_FILENAME = "postgresql.conf"  # Expected to be in the artifact

POSTGRES_MIN_CPU = 10
POSTGRES_MAX_CPU = 1000
POSTGRES_MIN_MEMORY = 32
POSTGRES_MAX_MEMORY = 1024

def run(
 plan,
 image="postgres:alpine",
 service_name="postgres",
 user="postgres",
 password="MyPassword1!",
 database="postgres",
 config_file_artifact_name="",
 seed_file_artifact_name="",
 extra_configs=[],
 persistent=True,
 launch_adminer=False,
 min_cpu=POSTGRES_MIN_CPU,
 max_cpu=POSTGRES_MAX_CPU,
 min_memory=POSTGRES_MIN_MEMORY,
 max_memory=POSTGRES_MAX_MEMORY,
 node_selectors=None,
):
	cmd = [] # 34
	files = {}
	env_vars = {
		"POSTGRES_DB": database,
		"POSTGRES_USER": user,
		"POSTGRES_PASSWORD": password,
	}
	
	if persistent:
		files[DATA_DIRECTORY_PATH] = Directory(
			persistent_key= "data-{0}".format(service_name),
		)
		env_vars["PGDATA"] = DATA_DIRECTORY_PATH + "/pgdata"
	if node_selectors == None:
		node_selectors = {}
	if config_file_artifact_name != "":
		config_filepath = CONFIG_FILE_MOUNT_DIRPATH + "/" + CONFIG_FILENAME
		cmd += ["-c", "config_file=" + config_filepath]
		files[CONFIG_FILE_MOUNT_DIRPATH] = config_file_artifact_name
	
	# append cmd with postgres config overrides passed by users
	if len(extra_configs) > 0:
		for config in extra_configs:
			cmd += ["-c", config]
	
	if seed_file_artifact_name != "":
		files[SEED_FILE_MOUNT_PATH] = seed_file_artifact_name
	
	postgres_service = plan.add_service(
		name=service_name,
		config=ServiceConfig(
			image=image,
			ports={
				PORT_NAME: PortSpec(
					number=5432,
					application_protocol=APPLICATION_PROTOCOL,
				)
			},
			cmd=cmd,
			files=files,
			env_vars=env_vars,
			min_cpu=min_cpu,
			max_cpu=max_cpu,
			min_memory=min_memory,
			max_memory=max_memory,
			node_selectors=node_selectors,
		),
	)
	
	url = "{protocol}://{user}:{password}@{hostname}/{database}".format(
		protocol=APPLICATION_PROTOCOL,
		user=user,
		password=password,
		hostname=postgres_service.hostname,
		database=database,
	)
	
	return struct(
		url=url,
		service=postgres_service,
		port=postgres_service.ports[PORT_NAME],
		user=user,
		password=password,
		database=database,
		min_cpu=min_cpu,
		max_cpu=max_cpu,
		min_memory=min_memory,
		max_memory=max_memory,
		node_selectors=node_selectors,
	)
`
	serializedJsonParams := "{}"
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), packageId, mainFunctionName, noPackageReplaceOptions, relativePathToMainFile, serializedPostgresPackageStarlark, serializedJsonParams, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		packageId,
		suite.packageContentProvider,
		noPackageReplaceOptions,
	)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString := `packageId: github.com/kurtosis-tech/postgres-package
services:
- name: postgres
  image: postgres:alpine
  envVars:
  - key: POSTGRES_USER
	value: postgres
  - key: POSTGRES_PASSWORD
    value: MyPassword1!
  - key: PGDATA
    value: /data//pgdata
  - key: POSTGRES_DB
    value: postgres
  ports:
  - name: postgresql
    number: 5432
    transportProtocol: TCP
    applicationProtocol: postgresql
  files:
	
`
	require.Equal(suite.T(), expectedYamlString, string(yamlBytes))
}

func (suite *PlanYamlGeneratorTestSuite) TestSimpleScriptWithFilesArtifact() {
	packageId := "github.com/kurtosis-tech/plan-yaml-prac"
	mainFunctionName := ""
	relativePathToMainFile := "main.star"

	serializedScript := `def run(plan, args):
    hi_files_artifact = plan.render_templates(
        config={
            "hi.txt":struct(
                template="hello world!",
                data={}
            )
        },
        name="hi-file"
    )

    plan.add_service(
        name="tedi",
        config=ServiceConfig(
            image="ubuntu:latest",
            cmd=["cat", "/root/hi.txt"],
            ports={
                "dashboard":PortSpec(
                    number=1234,
                    application_protocol="http",
                    transport_protocol="TCP"
                )
            },
            env_vars={
                "PASSWORD": "tedi"
            },
            files={
                "/root": hi_files_artifact,
            }
        )
    )
`
	serializedJsonParams := "{}"
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), packageId, mainFunctionName, noPackageReplaceOptions, relativePathToMainFile, serializedScript, serializedJsonParams, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	pyg := NewPlanYamlGenerator(
		instructionsPlan,
		suite.serviceNetwork,
		packageId,
		suite.packageContentProvider,
		noPackageReplaceOptions,
	)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString := `packageId: github.com/kurtosis-tech/postgres-package
services:
- name: postgres
  image: postgres:alpine
  envVars:
  - key: POSTGRES_USER
	value: postgres
  - key: POSTGRES_PASSWORD
    value: MyPassword1!
  - key: PGDATA
    value: /data//pgdata
  - key: POSTGRES_DB
    value: postgres
  ports:
  - name: postgresql
    number: 5432
    transportProtocol: TCP
    applicationProtocol: postgresql
  files:
	
`
	require.Equal(suite.T(), expectedYamlString, string(yamlBytes))
}

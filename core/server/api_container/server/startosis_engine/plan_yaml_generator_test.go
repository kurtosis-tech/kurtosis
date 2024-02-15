package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PlanYamlGeneratorTestSuite struct {
	suite.Suite
	serviceNetwork         *service_network.MockServiceNetwork
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
	runtimeValueStore      *runtime_value_store.RuntimeValueStore

	interpreter *StartosisInterpreter
}

func (suite *PlanYamlGeneratorTestSuite) SetupTest() {
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	enclaveDb := getEnclaveDBForTest(suite.T())

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(suite.T(), err)
	suite.runtimeValueStore = runtimeValueStore
	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "")

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

func TestRunPlanYamlGeneratorTestSuite(t *testing.T) {
	suite.Run(t, new(PlanYamlGeneratorTestSuite))
}

func (suite *PlanYamlGeneratorTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorVerySimpleScript() {
	script := `
def run(plan):

	service_name = "%v"

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

	pyg := NewPlanYamlGenerator(instructionsPlan)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	expectedYamlString :=
		`
packageId: DEFAULT_PACKAGE_ID_FOR_SCRIPT
services:
 name: 
 image: kurtosistech/example-datastore-server
 ports:
	name: grpc
	transportProtocol: TCP
	number: 1323
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

	pyg := NewPlanYamlGenerator(instructionsPlan)
	yamlBytes, err := pyg.GenerateYaml()
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), yamlBytes, []byte{})
}

func TestConvertPlanYamlToYamlBytes(t *testing.T) {
	PackageId := "github.com/kurtosis-tech/postgres-package"

	services := []*Service{
		{
			Name:  "tedi",
			Uuid:  "uuid",
			Image: "postgres:alpine",
			EnvVars: []*EnvironmentVariable{
				{
					Key:   "kevin",
					Value: "dag",
				},
			},
		},
		{
			Name:  "kaleb",
			Uuid:  "uuid",
			Image: "postgres:alpine",
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
			Uuid:  "something",
			Name:  "something",
			Files: nil,
		},
	}
	tasks := []*Task{
		{
			TaskType: PYTHON,
			Name:     "updateSomething",
			Command:  "do something",
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
	require.Equal(t, "", string(yamlBytes))
}

//func (suite *PlanYamlGeneratorTestSuite) TestPlanYamlGeneratorPostgresPackage() {
//	packageId := "github.com/kurtosis-tech/postgres-package"
//	mainFunctionName := ""
//	relativePathToMainFile := "main.star"
//	serializedStarlark :=
//		`adminer_module = import_module("github.com/bharath-123/db-adminer-package/main.star")
//PORT_NAME = "postgresql"
//APPLICATION_PROTOCOL = "postgresql"
//PG_DRIVER = "pgsql"
//
//CONFIG_FILE_MOUNT_DIRPATH = "/config"
//SEED_FILE_MOUNT_PATH = "/docker-entrypoint-initdb.d"
//DATA_DIRECTORY_PATH = "/data/"
//
//CONFIG_FILENAME = "postgresql.conf"  # Expected to be in the artifact
//
//POSTGRES_MIN_CPU = 10
//POSTGRES_MAX_CPU = 1000
//POSTGRES_MIN_MEMORY = 32
//POSTGRES_MAX_MEMORY = 1024
//
//	def run(
//	 plan,
//	 image="postgres:alpine",
//	 service_name="postgres",
//	 user="postgres",
//	 password="MyPassword1!",
//	 database="postgres",
//	 config_file_artifact_name="",
//	 seed_file_artifact_name="",
//	 extra_configs=[],
//	 persistent=True,
//	 launch_adminer=False,
//	 min_cpu=POSTGRES_MIN_CPU,
//	 max_cpu=POSTGRES_MAX_CPU,
//	 min_memory=POSTGRES_MIN_MEMORY,
//	 max_memory=POSTGRES_MAX_MEMORY,
//	 node_selectors=None,
//	):
//     """Launches a Postgresql database instance, optionally seeding it with a SQL file script
//
//     Args:
//         image (string): The container image that the Postgres service will be started with
//         service_name (string): The name to give the Postgres service
//         user (string): The user to create the Postgres database with
//         password (string): The password to give to the created user
//         database (string): The name of the database to create
//         config_file_artifact_name (string): The name of a files artifact that contains a Postgres config file in it
//             If not empty, this will be used to configure the Postgres server
//         seed_file_artifact_name (string): The name of a files artifact containing seed data
//             If not empty, the Postgres server will be populated with the data upon start
//         extra_configs (list[string]): Each argument gets passed as a '-c' argument to the Postgres server
//         persistent (bool): Whether the data should be persisted. Defaults to True; Note that this isn't supported on multi node k8s cluster as of 2023-10-16
//         launch_adminer (bool): Whether to launch adminer which launches a website to inspect postgres database entries. Defaults to False.
//         min_cpu (int): Define how much CPU millicores the service should be assigned at least.
//         max_cpu (int): Define how much CPU millicores the service should be assign max.
//         min_memory (int): Define how much MB of memory the service should be assigned at least.
//         max_memory (int): Define how much MB of memory the service should be assigned max.
//         node_selectors (dict[string, string]): Define a dict of node selectors - only works in kubernetes example: {"kubernetes.io/hostname": node-name-01}
//     Returns:
//         An object containing useful information about the Postgres database running inside the enclave:
//		{
//			"database": "postgres",
//			"password": "MyPassword1!",
//			"port": {
//				"application_protocol": "postgresql",
//					"number": 5432,
//					"transport_protocol": "TCP",
//					"wait": "2m0s"
//			},
//			"service": {
//				"hostname": "postgres",
//					"ip_address": "172.16.0.4",
//					"name": "postgres",
//					"ports": {
//					"postgresql": {
//						"application_protocol": "postgresql",
//							"number": 5432,
//							"transport_protocol": "TCP",
//							"wait": "2m0s"
//					}
//				}
//			},
//			"url": "postgresql://postgres:MyPassword1!@postgres/postgres",
//			"user": "postgres"
//		}
//	 """
//     cmd = []
//     files = {}
//     env_vars = {
//         "POSTGRES_DB": database,
//         "POSTGRES_USER": user,
//         "POSTGRES_PASSWORD": password,
//     }
//
//     if persistent:
//         files[DATA_DIRECTORY_PATH] = Directory(
//             persistent_key= "data-{0}".format(service_name),
//         )
//         env_vars["PGDATA"] = DATA_DIRECTORY_PATH + "/pgdata"
//     if node_selectors == None:
//         node_selectors = {}
//     if config_file_artifact_name != "":
//         config_filepath = CONFIG_FILE_MOUNT_DIRPATH + "/" + CONFIG_FILENAME
//         cmd += ["-c", "config_file=" + config_filepath]
//         files[CONFIG_FILE_MOUNT_DIRPATH] = config_file_artifact_name
//
//     # append cmd with postgres config overrides passed by users
//     if len(extra_configs) > 0:
//         for config in extra_configs:
//             cmd += ["-c", config]
//
//     if seed_file_artifact_name != "":
//         files[SEED_FILE_MOUNT_PATH] = seed_file_artifact_name
//
//     postgres_service = plan.add_service(
//         name=service_name,
//         config=ServiceConfig(
//             image=image,
//             ports={
//                 PORT_NAME: PortSpec(
//                     number=5432,
//                     application_protocol=APPLICATION_PROTOCOL,
//                 )
//             },
//             cmd=cmd,
//             files=files,
//             env_vars=env_vars,
//             min_cpu=min_cpu,
//             max_cpu=max_cpu,
//             min_memory=min_memory,
//             max_memory=max_memory,
//             node_selectors=node_selectors,
//         ),
//     )
//
//     if launch_adminer:
//         adminer = adminer_module.run(
//             plan,
//             default_db=database,
//             default_driver=PG_DRIVER,
//             default_password=password,
//             default_server=postgres_service.hostname,
//             default_username=user,
//         )
//
//     url = "{protocol}://{user}:{password}@{hostname}/{database}".format(
//         protocol=APPLICATION_PROTOCOL,
//         user=user,
//         password=password,
//         hostname=postgres_service.hostname,
//         database=database,
//     )
//
//     return struct(
//         url=url,
//         service=postgres_service,
//         port=postgres_service.ports[PORT_NAME],
//         user=user,
//         password=password,
//         database=database,
//         min_cpu=min_cpu,
//         max_cpu=max_cpu,
//         min_memory=min_memory,
//         max_memory=max_memory,
//         node_selectors=node_selectors,
//     )
//
// def run_query(plan, service, user, password, database, query):
//     url = "{protocol}://{user}:{password}@{hostname}/{database}".format(
//         protocol=APPLICATION_PROTOCOL,
//         user=user,
//         password=password,
//         hostname=service.hostname,
//         database=database,
//     )
//     return plan.exec(
//         service.name, recipe=ExecRecipe(command=["psql", url, "-c", query])
//     )
//`
//	serializedJsonParams := ""
//	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), packageId, mainFunctionName, noPackageReplaceOptions, relativePathToMainFile, serializedStarlark, serializedJsonParams, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask)
//	require.Nil(suite.T(), interpretationError)
//	require.Equal(suite.T(), 8, instructionsPlan.Size())
//
//	//pyg := NewPlanYamlGenerator(instructionsPlan)
//	//yamlBytes, err := pyg.GenerateYaml()
//	//require.NoError(suite.T(), err)
//	//
//	//require.Equal(suite.T(), yamlBytes, []byte{})
//}

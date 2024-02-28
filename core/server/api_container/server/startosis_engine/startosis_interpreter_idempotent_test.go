package startosis_engine

import (
	"context"
	"github.com/google/uuid"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"net"
	"testing"
)

const (
	starlarkValueSerdeThreadName = "test-serde-thread"

	enclaveUuid = enclave.EnclaveUUID("enclave-uuid")

	noInputParams = "{}"

	defaultNonBlockingMode = false
)

var noPackageReplaceOptions = map[string]string{}

type StartosisInterpreterIdempotentTestSuite struct {
	suite.Suite
	packageContentProvider *mock_package_content_provider.MockPackageContentProvider
	interpreter            *StartosisInterpreter
}

func (suite *StartosisInterpreterIdempotentTestSuite) SetupTest() {
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	enclaveDb := getEnclaveDBForTest(suite.T())

	thread := &starlark.Thread{
		Name:       starlarkValueSerdeThreadName,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make),

		kurtosis_types.ServiceTypeName: starlark.NewBuiltin(kurtosis_types.ServiceTypeName, kurtosis_types.NewServiceType().CreateBuiltin()),
		port_spec.PortSpecTypeName:     starlark.NewBuiltin(port_spec.PortSpecTypeName, port_spec.NewPortSpecType().CreateBuiltin()),
	}
	starlarkValueSerde := kurtosis_types.NewStarlarkValueSerde(thread, starlarkEnv)

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(starlarkValueSerde, enclaveDb)
	require.NoError(suite.T(), err)

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, starlarkValueSerde)
	require.Nil(suite.T(), err)

	serviceNetwork := service_network.NewMockServiceNetwork(suite.T())
	serviceNetwork.EXPECT().GetApiContainerInfo().Maybe().Return(
		service_network.NewApiContainerInfo(net.IPv4(0, 0, 0, 0), uint16(1234), "0.0.0"),
	)
	serviceNetwork.EXPECT().GetEnclaveUuid().Maybe().Return(enclaveUuid)
	suite.interpreter = NewStartosisInterpreter(serviceNetwork, suite.packageContentProvider, runtimeValueStore, starlarkValueSerde, "", interpretationTimeValueStore)
}

func TestRunStartosisInterpreterIdempotentTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisInterpreterIdempotentTestSuite))
}

func (suite *StartosisInterpreterIdempotentTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

// Most simple case - replay the same package twice
// Current plan ->     [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that this results in the entire set of instruction being skipped
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_IdenticalPackage() {
	script := `def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 3, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		script,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), `print(msg="instruction1")`, scheduledInstruction1.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), `print(msg="instruction2")`, scheduledInstruction2.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction3.IsExecuted())
}

// Add an instruction at the end of a package that was already run
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`]
// Check that the first two instructions are already executed, and the last one is in the new plan marked as not
// executed
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_AppendNewInstruction() {
	initialScript := `def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		initialScript,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 2, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	updatedScript := `def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
	plan.print("instruction3")
`
	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		updatedScript,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 0, instructionsPlan.GetIndexOfFirstInstruction())
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), `print(msg="instruction1")`, scheduledInstruction1.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), `print(msg="instruction2")`, scheduledInstruction2.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsExecuted())
}

// Run an instruction inside an enclave that is not empty (other non-related package were run in the past)
// Current plan ->     [`print("instruction1")`  `print("instruction2")`                         ]
// Package to run ->   [                                                  `print("instruction3")`]
// Check that the first two instructions are marked as imported from a previous plan, already executed, and the last
// one is in the new plan marked as not executed
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_DisjointInstructionSet() {
	initialScript := `def run(plan, args):
	plan.print("instruction1")
	plan.print("instruction2")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		initialScript,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 2, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	updatedScript := `def run(plan, args):
	plan.print("instruction3")
`
	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		updatedScript,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 2, instructionsPlan.GetIndexOfFirstInstruction()) // the first 2 instruction are kept
	require.Equal(suite.T(), 1, len(instructionSequence))                      // the new one is the only one inside the plan

	scheduledInstruction3 := instructionSequence[0]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsExecuted())
}

// Submit a package with an update on an instruction located "in the middle" of the package
// Current plan ->     [`print("instruction1")`  `print("instruction2")`      `print("instruction3")`]
// Package to run ->   [`print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// That will result in the concatenation of the two plans because we're not able to properly resolve dependencies yet
// [`print("instruction1")`  `print("instruction2")`  `print("instruction3")`  `print("instruction1")`  `print("instruction2_NEW")`  `print("instruction3")`]
// The first three are imported from a previous plan, already executed, while the last three are from all new
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_InvalidNewVersionOfThePackage() {
	initialScript := `def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2")
	plan.print(msg="instruction3")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		initialScript,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 3, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	updatedScript := `def run(plan, args):
	plan.print(msg="instruction1")
	plan.print(msg="instruction2_NEW")
	plan.print(msg="instruction3")
`
	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		updatedScript,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 3, instructionsPlan.GetIndexOfFirstInstruction())
	require.Equal(suite.T(), 3, len(instructionSequence))

	scheduledInstruction4 := instructionSequence[0]
	require.Equal(suite.T(), `print(msg="instruction1")`, scheduledInstruction4.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction4.IsExecuted())

	scheduledInstruction5 := instructionSequence[1]
	require.Equal(suite.T(), `print(msg="instruction2_NEW")`, scheduledInstruction5.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction5.IsExecuted())

	scheduledInstruction6 := instructionSequence[2]
	require.Equal(suite.T(), `print(msg="instruction3")`, scheduledInstruction6.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction6.IsExecuted())
}

// Submit a package with an update to an add_service instruction. add_service instructions is supports being run twice
// in an enclave, updating "live" the service underneath. Check that the add_service and its direct dependencies are
// scheduled for a re-run but other instructions remains "SKIPPED"
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_AddServiceIdempotency() {
	initialScript := `def run(plan):
	service_1 = plan.add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.2.3"))
	plan.print("Service 1 - IP: {} - Hostname: {}".format(service_1.ip_address, service_1.hostname))
	plan.exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"]))
	plan.verify(value=service_1.ip_address, assertion="==", target_value="fake_ip")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		initialScript,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 4, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	updatedScript := `def run(plan):
	service_1 = plan.add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.5.0")) # <-- version updated
	plan.print("Service 1 - IP: {} - Hostname: {}".format(service_1.ip_address, service_1.hostname)) # <-- identical
	plan.exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"])) # <-- identical but should be rerun b/c service_1 updated
	plan.verify(value=service_1.ip_address, assertion="==", target_value="fake_ip") # <-- identical b/c we don't track runtime value provenance yet
`
	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		updatedScript,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 0, instructionsPlan.GetIndexOfFirstInstruction())
	require.Equal(suite.T(), 4, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), `add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.5.0"))`, scheduledInstruction1.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Regexp(suite.T(), `print\(msg="Service 1 - IP: {{kurtosis:[a-z0-9]{32}:ip_address\.runtime_value}} - Hostname: {{kurtosis:[a-z0-9]{32}:hostname\.runtime_value}}"\)`, scheduledInstruction2.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction2.IsExecuted())

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"]))`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsExecuted())

	scheduledInstruction4 := instructionSequence[3]
	require.Regexp(suite.T(), `verify\(value="{{kurtosis:[a-z0-9]{32}:ip_address\.runtime_value}}", assertion="==", target_value="fake_ip"\)`, scheduledInstruction4.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction4.IsExecuted())
}

// Submit a package with an update to an upload_files instruction. upload_files instructions support being run twice
// in an enclave, updating "live" the file underneath if it has changed. Check that the add_service and its direct
// dependencies are scheduled for a re-run but other instructions remains "SKIPPED"
func (suite *StartosisInterpreterIdempotentTestSuite) TestInterpretAndOptimize_UploadFilesIdempotency() {
	initialScript := `def run(plan):
	files_artifact = plan.render_templates(
        name="files_artifact",
        config={
            "/output.txt": struct(template="Hello {{.World}}", data={"World" : "World!"}),
        }
    )
	service_1 = plan.add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.2.3", files={"/path/": files_artifact}))
	plan.exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"]))
	plan.verify(value=service_1.ip_address, assertion="==", target_value="fake_ip")
`
	// Interpretation of the initial script to generate the current enclave plan
	_, currentEnclavePlan, interpretationApiErr := suite.interpreter.Interpret(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		useDefaultMainFunctionName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		initialScript,
		noInputParams,
		defaultNonBlockingMode,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0))
	require.Nil(suite.T(), interpretationApiErr)
	require.Equal(suite.T(), 4, currentEnclavePlan.Size())
	convertedEnclavePlan := suite.convertInstructionPlanToEnclavePlan(currentEnclavePlan)

	updatedScript := `def run(plan):
	files_artifact = plan.render_templates(
        name="files_artifact",
        config={
            "/output.txt": struct(template="Bonjour {{.World}}", data={"World" : "World!"}), # updated template!!
        }
    )
	service_1 = plan.add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.2.3", files={"/path/": files_artifact}))
	plan.exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"]))
	plan.verify(value=service_1.ip_address, assertion="==", target_value="fake_ip")
`
	// Interpret the updated script against the current enclave plan
	_, instructionsPlan, interpretationError := suite.interpreter.InterpretAndOptimizePlan(
		context.Background(),
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		useDefaultMainFunctionName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		updatedScript,
		noInputParams,
		defaultNonBlockingMode,
		convertedEnclavePlan,
	)
	require.Nil(suite.T(), interpretationError)

	instructionSequence, err := instructionsPlan.GeneratePlan()
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), 0, instructionsPlan.GetIndexOfFirstInstruction())
	require.Equal(suite.T(), 4, len(instructionSequence))

	scheduledInstruction1 := instructionSequence[0]
	require.Equal(suite.T(), `render_templates(config={"/output.txt": struct(data={"World": "World!"}, template="Bonjour {{.World}}")}, name="files_artifact")`, scheduledInstruction1.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction1.IsExecuted())

	scheduledInstruction2 := instructionSequence[1]
	require.Equal(suite.T(), `add_service(name="service_1", config=ServiceConfig(image="kurtosistech/image:1.2.3", files={"/path/": "files_artifact"}))`, scheduledInstruction2.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction2.IsExecuted()) // since the files artifact has been updated, the service will be updated as well

	scheduledInstruction3 := instructionSequence[2]
	require.Equal(suite.T(), `exec(service_name="service_1", recipe=ExecRecipe(command=["echo", "Hello World!"]))`, scheduledInstruction3.GetInstruction().String())
	require.False(suite.T(), scheduledInstruction3.IsExecuted()) // since the service has been updated, the exec will be re-run

	scheduledInstruction4 := instructionSequence[3]
	require.Regexp(suite.T(), `verify\(value="{{kurtosis:[a-z0-9]{32}:ip_address\.runtime_value}}", assertion="==", target_value="fake_ip"\)`, scheduledInstruction4.GetInstruction().String())
	require.True(suite.T(), scheduledInstruction4.IsExecuted()) // this instruction is not affected, i.e. it won't be re-run
}

func (suite *StartosisInterpreterIdempotentTestSuite) convertInstructionPlanToEnclavePlan(instructionPlan *instructions_plan.InstructionsPlan) *enclave_plan_persistence.EnclavePlan {
	enclavePlan := enclave_plan_persistence.NewEnclavePlan()
	instructionPlanSequence, interpretationErr := instructionPlan.GeneratePlan()
	suite.Require().Nil(interpretationErr)
	for _, instruction := range instructionPlanSequence {
		enclavePlanInstruction, err := instruction.GetInstruction().GetPersistableAttributes().SetUuid(
			uuid.New().String(),
		).SetReturnedValue(
			"None", // the returnedValue does not matter for those tests as we're testing only the interpretation phase
		).Build()
		suite.Require().NoError(err)
		enclavePlan.AppendInstruction(enclavePlanInstruction)
	}
	return enclavePlan
}

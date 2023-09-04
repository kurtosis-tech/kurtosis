package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/request"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/start_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/stop_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/tasks"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/directory"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/time_module_with_now_disabled"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func Predeclared() starlark.StringDict {
	return starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark

		// go-starlark time module with time.now() disabled
		time_module_with_now_disabled.TimeNowBuiltinName: time_module_with_now_disabled.GetTimeModuleWithNowDisabled(),
	}
}

// KurtosisPlanInstructions returns the entire list of KurtosisPlanInstruction.
//
// A KurtosisPlanInstruction is a builtin that adds an operation to the plan.
// It is typically (but not only) a builtin that modify the state of the Kurtosis enclave. A KurtosisPlanInstruction
// can have an effect at both interpretation and execution time.
//
// Examples: add_service, exec, wait, etc.
func KurtosisPlanInstructions(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore, packageContentProvider startosis_packages.PackageContentProvider) []*kurtosis_plan_instruction.KurtosisPlanInstruction {
	return []*kurtosis_plan_instruction.KurtosisPlanInstruction{
		add_service.NewAddService(serviceNetwork, runtimeValueStore),
		add_service.NewAddServices(serviceNetwork, runtimeValueStore),
		assert.NewAssert(runtimeValueStore),
		exec.NewExec(serviceNetwork, runtimeValueStore),
		kurtosis_print.NewPrint(serviceNetwork, runtimeValueStore),
		remove_service.NewRemoveService(serviceNetwork),
		render_templates.NewRenderTemplatesInstruction(serviceNetwork, runtimeValueStore),
		request.NewRequest(serviceNetwork, runtimeValueStore),
		start_service.NewStartService(serviceNetwork),
		tasks.NewRunPythonService(serviceNetwork, runtimeValueStore),
		tasks.NewRunShService(serviceNetwork, runtimeValueStore),
		stop_service.NewStopService(serviceNetwork),
		store_service_files.NewStoreServiceFiles(serviceNetwork),
		upload_files.NewUploadFiles(serviceNetwork, packageContentProvider),
		wait.NewWait(serviceNetwork, runtimeValueStore),
	}
}

// KurtosisHelpers returns the entire list of KurtosisHelper available by default in Kurtosis Starlark engine.
//
// A KurtosisHelper is a builtin which provides additional capabilities at interpretation time. It does not have any
// effect at execution time. It can be thought as a Starlark builtin that could exist in a world without the
// Kurtosis enclave.
//
// Example: read_file, import_package, etc.
func KurtosisHelpers(recursiveInterpret func(moduleId string, scriptContent string) (starlark.StringDict, *startosis_errors.InterpretationError), packageContentProvider startosis_packages.PackageContentProvider, packageGlobalCache map[string]*startosis_packages.ModuleCacheEntry) []*starlark.Builtin {
	read_file.NewReadFileHelper(packageContentProvider)
	return []*starlark.Builtin{
		starlark.NewBuiltin(import_module.ImportModuleBuiltinName, import_module.NewImportModule(recursiveInterpret, packageContentProvider, packageGlobalCache).CreateBuiltin()),
		starlark.NewBuiltin(print_builtin.PrintBuiltinName, print_builtin.GeneratePrintBuiltin()),
		starlark.NewBuiltin(read_file.ReadFileBuiltinName, read_file.NewReadFileHelper(packageContentProvider).CreateBuiltin()),
	}
}

// KurtosisTypeConstructors returns the entire list of Kurtosis Starlark type constructors.
//
// A KurtosisTypeConstructor is a builtin that has for sole purpose to instantiate a Kurtosis specific object type
// (i.e. a constructor in the OOP language).
//
// Example: ServiceConfig, PortSpec, etc.
func KurtosisTypeConstructors() []*starlark.Builtin {
	return []*starlark.Builtin{
		starlark.NewBuiltin(kurtosis_types.ServiceTypeName, kurtosis_types.NewServiceType().CreateBuiltin()),
		starlark.NewBuiltin(directory.DirectoryTypeName, directory.NewDirectoryType().CreateBuiltin()),
		starlark.NewBuiltin(recipe.ExecRecipeTypeName, recipe.NewExecRecipeType().CreateBuiltin()),
		starlark.NewBuiltin(recipe.GetHttpRecipeTypeName, recipe.NewGetHttpRequestRecipeType().CreateBuiltin()),
		starlark.NewBuiltin(recipe.PostHttpRecipeTypeName, recipe.NewPostHttpRequestRecipeType().CreateBuiltin()),
		starlark.NewBuiltin(port_spec.PortSpecTypeName, port_spec.NewPortSpecType().CreateBuiltin()),
		starlark.NewBuiltin(service_config.ServiceConfigTypeName, service_config.NewServiceConfigType().CreateBuiltin()),
		starlark.NewBuiltin(service_config.ReadyConditionTypeName, service_config.NewReadyConditionType().CreateBuiltin()),
	}
}

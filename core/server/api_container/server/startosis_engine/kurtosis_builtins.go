package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_connection"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/request"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/set_connection"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/update_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
)

// KurtosisPlanInstructions returns the entire list of KurtosisPlanInstruction.
//
// A KurtosisPlanInstruction is a builtin that adds an operation to the plan.
// It is typically (but not only) a builtin that modify the state of the Kurtosis enclave. A KurtosisPlanInstruction
// can have an effect at both interpretation and execution time.
//
// Examples: add_service, exec, wait, etc.
func KurtosisPlanInstructions(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) []*kurtosis_plan_instruction.KurtosisPlanInstruction {
	return []*kurtosis_plan_instruction.KurtosisPlanInstruction{
		add_service.NewAddService(serviceNetwork, runtimeValueStore),
		add_service.NewAddServices(serviceNetwork, runtimeValueStore),
		assert.NewAssert(runtimeValueStore),
		exec.NewExec(serviceNetwork, runtimeValueStore),
		remove_service.NewRemoveService(serviceNetwork),
		render_templates.NewRenderTemplatesInstruction(serviceNetwork),
		request.NewRequest(serviceNetwork, runtimeValueStore),
		set_connection.NewSetConnection(serviceNetwork),
	}
}

func OldKurtosisPlanInstructions(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, packageContentProvider startosis_packages.PackageContentProvider, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) []*starlark.Builtin {
	return []*starlark.Builtin{
		starlark.NewBuiltin(kurtosis_print.PrintBuiltinName, kurtosis_print.GeneratePrintBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
		starlark.NewBuiltin(remove_connection.RemoveConnectionBuiltinName, remove_connection.GenerateRemoveConnectionBuiltin(instructionsQueue, serviceNetwork)),
		starlark.NewBuiltin(store_service_files.StoreServiceFilesBuiltinName, store_service_files.GenerateStoreServiceFilesBuiltin(instructionsQueue, serviceNetwork)),
		starlark.NewBuiltin(update_service.UpdateServiceBuiltinName, update_service.GenerateUpdateServiceBuiltin(instructionsQueue, serviceNetwork)),
		starlark.NewBuiltin(upload_files.UploadFilesBuiltinName, upload_files.GenerateUploadFilesBuiltin(instructionsQueue, packageContentProvider, serviceNetwork)),
		starlark.NewBuiltin(wait.WaitBuiltinName, wait.GenerateWaitBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
	}
}

// KurtosisHelpers returns the entire list of KurtosisHelper available by default in Kurtosis Starlark engine.
//
// A KurtosisHelper is a builtin which provides additional capabilities at interpretation time. It does not have any
// effect at execution time. It can be thought as a Starlark builtin that could exist in a world without the
// Kurtosis enclave.
//
// Example: read_file, import_package, etc.
func KurtosisHelpers(recursiveInterpret func(moduleId string, scriptContent string) (starlark.StringDict, error), packageContentProvider startosis_packages.PackageContentProvider, packageGlobalCache map[string]*startosis_packages.ModuleCacheEntry) []*starlark.Builtin {
	return []*starlark.Builtin{
		starlark.NewBuiltin(import_module.ImportModuleBuiltinName, import_module.GenerateImportBuiltin(recursiveInterpret, packageContentProvider, packageGlobalCache)),
		starlark.NewBuiltin(print_builtin.PrintBuiltinName, print_builtin.GeneratePrintBuiltin()),
		starlark.NewBuiltin(read_file.ReadFileBuiltinName, read_file.GenerateReadFileBuiltin(packageContentProvider)),
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
		starlark.NewBuiltin(recipe.ExecRecipeName, recipe.MakeExecRequestRecipe),
		starlark.NewBuiltin(recipe.GetHttpRecipeTypeName, recipe.MakeGetHttpRequestRecipe),
		starlark.NewBuiltin(recipe.PostHttpRecipeTypeName, recipe.MakePostHttpRequestRecipe),
		starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		starlark.NewBuiltin(kurtosis_types.PortSpecTypeName, kurtosis_types.MakePortSpec),
		starlark.NewBuiltin(kurtosis_types.ServiceConfigTypeName, kurtosis_types.MakeServiceConfig),
		starlark.NewBuiltin(kurtosis_types.UpdateServiceConfigTypeName, kurtosis_types.MakeUpdateServiceConfig),
		starlark.NewBuiltin(kurtosis_types.PacketDelayName, kurtosis_types.MakePacketDelay),
	}
}

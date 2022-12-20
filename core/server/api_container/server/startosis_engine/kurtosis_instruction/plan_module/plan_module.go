package plan_module

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
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
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	planModuleName = "plan"
)

func PlanModule(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore, contentProvider startosis_packages.PackageContentProvider) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: planModuleName,
		Members: starlark.StringDict{
			// Kurtosis instructions - will push instructions to the queue that will affect the enclave state at execution
			add_service.AddServiceBuiltinName:                starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, serviceNetwork, runtimeValueStore)),
			assert.AssertBuiltinName:                         starlark.NewBuiltin(assert.AssertBuiltinName, assert.GenerateAssertBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
			exec.ExecBuiltinName:                             starlark.NewBuiltin(exec.ExecBuiltinName, exec.GenerateExecBuiltin(instructionsQueue, serviceNetwork, runtimeValueStore)),
			kurtosis_print.PrintBuiltinName:                  starlark.NewBuiltin(kurtosis_print.PrintBuiltinName, kurtosis_print.GeneratePrintBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
			remove_connection.RemoveConnectionBuiltinName:    starlark.NewBuiltin(remove_connection.RemoveConnectionBuiltinName, remove_connection.GenerateRemoveConnectionBuiltin(instructionsQueue, serviceNetwork)),
			remove_service.RemoveServiceBuiltinName:          starlark.NewBuiltin(remove_service.RemoveServiceBuiltinName, remove_service.GenerateRemoveServiceBuiltin(instructionsQueue, serviceNetwork)),
			render_templates.RenderTemplatesBuiltinName:      starlark.NewBuiltin(render_templates.RenderTemplatesBuiltinName, render_templates.GenerateRenderTemplatesBuiltin(instructionsQueue, serviceNetwork)),
			request.RequestBuiltinName:                       starlark.NewBuiltin(request.RequestBuiltinName, request.GenerateRequestBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
			set_connection.SetConnectionBuiltinName:          starlark.NewBuiltin(set_connection.SetConnectionBuiltinName, set_connection.GenerateSetConnectionBuiltin(instructionsQueue, serviceNetwork)),
			store_service_files.StoreServiceFilesBuiltinName: starlark.NewBuiltin(store_service_files.StoreServiceFilesBuiltinName, store_service_files.GenerateStoreServiceFilesBuiltin(instructionsQueue, serviceNetwork)),
			update_service.UpdateServiceBuiltinName:		  starlark.NewBuiltin(update_service.UpdateServiceBuiltinName, update_service.GenerateUpdateServiceBuiltin(instructionsQueue, serviceNetwork)),
			upload_files.UploadFilesBuiltinName:              starlark.NewBuiltin(upload_files.UploadFilesBuiltinName, upload_files.GenerateUploadFilesBuiltin(instructionsQueue, contentProvider, serviceNetwork)),
			wait.WaitBuiltinName:                             starlark.NewBuiltin(wait.WaitBuiltinName, wait.GenerateWaitBuiltin(instructionsQueue, runtimeValueStore, serviceNetwork)),
		},
	}
}

/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package modules

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/stacktrace"
)

type ModuleID string

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
type ModuleContext struct {
	client   kurtosis_core_rpc_api_bindings.ApiContainerServiceClient
	moduleId ModuleID
}

func NewModuleContext(client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, moduleId ModuleID) *ModuleContext {
	return &ModuleContext{client: client, moduleId: moduleId}
}

// Docs available at https://docs.kurtosistech.com/kurtosis-client/lib-documentation
func (moduleCtx *ModuleContext) Execute(serializedParams string) (serializedResult string, resultErr error) {
	args := binding_constructors.NewExecuteModuleArgs(string(moduleCtx.moduleId), serializedParams)
	resp, err := moduleCtx.client.ExecuteModule(context.Background(), args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing module '%v'", moduleCtx.moduleId)
	}
	return resp.SerializedResult, nil
}


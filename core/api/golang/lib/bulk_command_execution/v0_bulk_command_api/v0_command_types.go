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

package v0_bulk_command_api

import "github.com/kurtosis-tech/stacktrace"

// We provide a visitor interface here so that:
//  1) all enum cases can be exhaustively handled
//  2) any changes in the enum will result in a compile break
type V0CommandTypeVisitor interface {
	VisitLoadModule() error
	VisitExecuteModule() error
	VisitRegisterService() error
	VisitStartService() error
	VisitRemoveService() error
	VisitRepartition() error
	VisitExecCommand() error
	VisitWaitForHttpGetEndpointAvailability() error
	VisitWaitForHttpPostEndpointAvailability() error
	VisitExecuteBulkCommands() error
	VisitGetServices() error
	VisitGetModules() error
}

type V0CommandType string

const (
	// vvvvvvvvvvvvvvvvvvvv Update the visitor whenever you add an enum value!!! vvvvvvvvvvvvvvvvvvvvvvvvvvv
	LoadModuleCommandType                          V0CommandType = "LOAD_MODULE"
	ExecuteModuleCommandType                       V0CommandType = "EXECUTE_MODULE"
	RegisterServiceCommandType                     V0CommandType = "REGISTER_SERVICE"
	StartServiceCommandType                        V0CommandType = "START_SERVICE"
	RemoveServiceCommandType                       V0CommandType = "REMOVE_SERVICE"
	RepartitionCommandType                         V0CommandType = "REPARTITION"
	ExecCommandCommandType                         V0CommandType = "EXEC_COMMAND"
	WaitForHttpGetEndpointAvailabilityCommandType  V0CommandType = "WAIT_FOR_HTTP_GET_ENDPOINT_AVAILABILITY"
	WaitForHttpPostEndpointAvailabilityCommandType V0CommandType = "WAIT_FOR_HTTP_POST_ENDPOINT_AVAILABILITY"
	ExecuteBulkCommandsCommandType                 V0CommandType = "EXECUTE_BULK_COMMANDS"
	GetServicesCommandType                         V0CommandType = "GET_SERVICES"
	GetModulesCommandType                          V0CommandType = "GET_MODULES"
	// ^^^^^^^^^^^^^^^^^^^^ Update the visitor whenever you add an enum value!!! ^^^^^^^^^^^^^^^^^^^^^^^^^^^
)

func (commandType V0CommandType) AcceptVisitor(visitor V0CommandTypeVisitor) error {
	var err error
	switch commandType {
	case LoadModuleCommandType:
		err = visitor.VisitLoadModule()
	case ExecuteModuleCommandType:
		err = visitor.VisitExecuteModule()
	case RegisterServiceCommandType:
		err = visitor.VisitRegisterService()
	case StartServiceCommandType:
		err = visitor.VisitStartService()
	case RemoveServiceCommandType:
		err = visitor.VisitRemoveService()
	case RepartitionCommandType:
		err = visitor.VisitRepartition()
	case ExecCommandCommandType:
		err = visitor.VisitExecCommand()
	case WaitForHttpGetEndpointAvailabilityCommandType:
		err = visitor.VisitWaitForHttpGetEndpointAvailability()
	case WaitForHttpPostEndpointAvailabilityCommandType:
		err = visitor.VisitWaitForHttpPostEndpointAvailability()
	case ExecuteBulkCommandsCommandType:
		err = visitor.VisitExecuteBulkCommands()
	case GetServicesCommandType:
		err = visitor.VisitGetServices()
	case GetModulesCommandType:
		err = visitor.VisitGetModules()
	default:
		return stacktrace.NewError("Unrecognized command type '%v'", commandType)
	}
	if err != nil {
		return stacktrace.Propagate(err, "The visitor returned an error for command type '%v'", commandType)
	}
	return nil
}

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

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ====================================================================================================
//                                   Command Arg Deserialization Visitor
// ====================================================================================================

// Visitor that will be used to deserialize command args into
type cmdArgDeserializingVisitor struct {
	bytesToDeserialize         []byte
	deserializedCommandArgsPtr proto.Message
}

func newCmdArgDeserializingVisitor(bytesToDeserialize []byte) *cmdArgDeserializingVisitor {
	return &cmdArgDeserializingVisitor{bytesToDeserialize: bytesToDeserialize}
}

func (visitor *cmdArgDeserializingVisitor) VisitLoadModule() error {
	args := &kurtosis_core_rpc_api_bindings.LoadModuleArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the module-loading args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitExecuteModule() error {
	args := &kurtosis_core_rpc_api_bindings.ExecuteModuleArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the module-executing args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitRegisterService() error {
	args := &kurtosis_core_rpc_api_bindings.RegisterServiceArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the register service args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitStartService() error {
	args := &kurtosis_core_rpc_api_bindings.StartServiceArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the start service args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitRemoveService() error {
	args := &kurtosis_core_rpc_api_bindings.RemoveServiceArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the remove service args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitRepartition() error {
	args := &kurtosis_core_rpc_api_bindings.RepartitionArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the repartition service args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitExecCommand() error {
	args := &kurtosis_core_rpc_api_bindings.ExecCommandArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the exec command args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitWaitForHttpGetEndpointAvailability() error {
	args := &kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the endpoint availability-waiting-http-get args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitWaitForHttpPostEndpointAvailability() error {
	args := &kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the endpoint availability-waiting-http-post args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitExecuteBulkCommands() error {
	args := &kurtosis_core_rpc_api_bindings.ExecuteBulkCommandsArgs{}
	if err := json.Unmarshal(visitor.bytesToDeserialize, args); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the bulk command execution args")
	}
	visitor.deserializedCommandArgsPtr = args
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitGetServices() error {
	visitor.deserializedCommandArgsPtr = &emptypb.Empty{}
	return nil
}

func (visitor *cmdArgDeserializingVisitor) VisitGetModules() error {
	visitor.deserializedCommandArgsPtr = &emptypb.Empty{}
	return nil
}

func (visitor *cmdArgDeserializingVisitor) GetDeserializedCommandArgs() proto.Message {
	return visitor.deserializedCommandArgsPtr
}

// ====================================================================================================
//                                        Serializable Command
// ====================================================================================================

// Used for serializing
type V0SerializableCommand struct {
	Type V0CommandType `json:"type"`

	// The only allowed objects here are from the bindings generated from the .proto file
	ArgsPtr proto.Message `json:"args"`
}

// A V0SerializableCommand knows how to deserialize itself, thanks to the "type" tag
func (cmd *V0SerializableCommand) UnmarshalJSON(bytes []byte) error {
	interstitialStruct := struct {
		Type      V0CommandType   `json:"type"`
		ArgsBytes json.RawMessage `json:"args"`
	}{}
	if err := json.Unmarshal(bytes, &interstitialStruct); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the bytes into a command")
	}

	visitor := newCmdArgDeserializingVisitor(interstitialStruct.ArgsBytes)
	if err := interstitialStruct.Type.AcceptVisitor(visitor); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing command with the following JSON:\n%v", string(bytes))
	}

	cmd.Type = interstitialStruct.Type
	cmd.ArgsPtr = visitor.GetDeserializedCommandArgs()

	return nil
}


// ====================================================================================================
//                                   Bulk Commands Package
// ====================================================================================================

type V0BulkCommands struct {
	Commands []V0SerializableCommand `json:"commands"`
}

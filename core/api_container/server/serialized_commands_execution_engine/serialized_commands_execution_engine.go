/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package json_parser

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_rpc_api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
)

type CommandType string

const (
	RegisterServiceCommandType CommandType = "REGISTER_SERVICE"
	StartServiceCommandType = "START_SERVICE"
)

type V0JsonDocument struct {
	Commands	[]GenericCommand	`json:"commands"`
}

type GenericCommand struct {
	Type CommandType		`json:"type"`
	Args json.RawMessage	`json:"args"`
}

type SerializedCommandsExecutionEngine struct {
	serviceNetwork *service_network.ServiceNetwork
}

func (engine *SerializedCommandsExecutionEngine) ExecuteSerializedCommands(schemaVersion bindings.ExecuteSerializedCommandsArgs_SchemaVersion, jsonStr string) error {
	switch schemaVersion {
	case bindings.ExecuteSerializedCommandsArgs_V0:
		parsedDocument := new(V0JsonDocument)
		if err := json.Unmarshal([]byte(jsonStr), &parsedDocument); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing the v%v JSON document", schemaVersion)
		}

		for idx, genericCommand := range parsedDocument.Commands {
			if err := parseAndExecuteCommand(engine.serviceNetwork, genericCommand); err != nil {
				return stacktrace.Propagate(err, "An error occurred parsing & executing command #%v", idx)
			}
		}
	default:
		return stacktrace.NewError("Encountered unrecognized schema version '%v'", schemaVersion)
	}

	return nil
}

func parseAndExecuteCommand(serviceNetwork *service_network.ServiceNetwork, command GenericCommand) error {
	// TODO run pre-parser to translate service ID -> IP addresses

	switch command.Type {
	case RegisterServiceCommandType:
		registerServiceArgs := new(bindings.RegisterServiceArgs)
		if err := json.Unmarshal(command.Args, &registerServiceArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing register service args JSON '%v'", string(command.Args))
		}

		if _, err := serviceNetwork.RegisterService(
			service_network_types.ServiceID(registerServiceArgs.ServiceId),
			service_network_types.PartitionID(registerServiceArgs.PartitionId),
		); err != nil {
			return stacktrace.Propagate(err, "An error occurred registering the service with the service network")
		}
	default:
		return stacktrace.NewError("Unrecognized command type '%v'", command.Type)
	}

	return nil
}
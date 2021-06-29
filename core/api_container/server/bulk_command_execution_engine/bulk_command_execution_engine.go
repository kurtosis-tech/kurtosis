/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package bulk_command_execution_engine

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
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

type BulkCommandExecutionEngine struct {
	serviceNetworkProxy *server.ServiceNetworkRpcApiProxy
	serviceNetwork *service_network.ServiceNetwork
}


func (engine *BulkCommandExecutionEngine) ExecuteCommands(schemaVersion core_api_bindings.BulkExecuteCommandsArgs_SchemaVersion, jsonStr string) error {
	switch schemaVersion {
	case core_api_bindings.BulkExecuteCommandsArgs_V0:
		parsedDocument := new(V0JsonDocument)
		if err := json.Unmarshal([]byte(jsonStr), &parsedDocument); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing the v%v JSON document", schemaVersion)
		}

		for idx, genericCommand := range parsedDocument.Commands {
			if err := parseAndExecuteCommand(engine.serviceNetwork, genericCommand); err != nil {
				return stacktrace.Propagate(err, "An error occurred parsing and executing command #%v", idx)
			}
		}
	default:
		return stacktrace.NewError("Encountered unrecognized commands schema version '%v'", schemaVersion)
	}

	return nil
}

func parseAndExecuteCommand(serviceNetwork *service_network.ServiceNetwork, command GenericCommand) error {
	// TODO run pre-parser to translate service ID -> IP addresses

	switch command.Type {
	case RegisterServiceCommandType:
		registerServiceArgs := new(core_api_bindings.RegisterServiceArgs)
		if err := json.Unmarshal(command.Args, &registerServiceArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing register service args JSON '%v'", string(command.Args))
		}

		serviceId := service_network_types.ServiceID(registerServiceArgs.ServiceId)
		if _, err := serviceNetwork.RegisterService(
			serviceId,
			service_network_types.PartitionID(registerServiceArgs.PartitionId),
		); err != nil {
			return stacktrace.Propagate(err, "An error occurred registering new service with service ID '%v' with the service network", serviceId)
		}
	case StartServiceCommandType:
		startServiceArgs := new(core_api_bindings.StartServiceArgs)
		if err := json.Unmarshal(command.Args, &startServiceArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing start service args JSON '%v'", string(command.Args))
		}

		serviceId := service_network_types.ServiceID(startServiceArgs.ServiceId)
		serviceNetwork.StartService(
			context.Background(),
			serviceId,
			startServiceArgs.DockerImage,
			startServiceArgs.)

		serviceId := service_network_types.ServiceID(registerServiceArgs.ServiceId)
		if _, err := serviceNetwork.RegisterService(
			serviceId,
			service_network_types.PartitionID(registerServiceArgs.PartitionId),
		); err != nil {
			return stacktrace.Propagate(err, "An error occurred registering new service with service ID '%v' with the service network", serviceId)
		}
	default:
		return stacktrace.NewError("Unrecognized command type '%v'", command.Type)
	}

	return nil
}
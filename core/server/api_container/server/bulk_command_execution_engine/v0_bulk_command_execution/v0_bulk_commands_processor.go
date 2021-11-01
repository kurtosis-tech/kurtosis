/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package v0_bulk_command_execution

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/bulk_command_execution/v0_bulk_command_api"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/bulk_command_execution_engine/service_ip_replacer"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network"
	"github.com/palantir/stacktrace"
)

type V0BulkCommandProcessor struct {
	apiService kurtosis_core_rpc_api_bindings.ApiContainerServiceServer
	ipReplacer *service_ip_replacer.ServiceIPReplacer
}

func NewV0BulkCommandProcessor(serviceNetwork service_network.ServiceNetwork, apiService kurtosis_core_rpc_api_bindings.ApiContainerServiceServer) (*V0BulkCommandProcessor, error) {
	ipReplacer, err := service_ip_replacer.NewServiceIPReplacer(
		v0_bulk_command_api.ServiceIdIpReplacementPrefix,
		v0_bulk_command_api.ServiceIdIpReplacementSuffix,
		serviceNetwork,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the v0 service ID -> IP replacer")
	}
	return &V0BulkCommandProcessor{
		apiService: apiService,
		ipReplacer: ipReplacer,
	}, nil
}

func (processor *V0BulkCommandProcessor) Process(ctx context.Context, serializedDocumentBody []byte) error {
	deserialized := new(v0_bulk_command_api.V0BulkCommands)
	if err := json.Unmarshal(serializedDocumentBody, &deserialized); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the bulk commands object")
	}

	for idx, command := range deserialized.Commands {
		cmdProcessingVisitor := newV0CommandProcessingVisitor(ctx, command.ArgsPtr, processor.ipReplacer, processor.apiService)
		if err := command.Type.AcceptVisitor(cmdProcessingVisitor); err != nil {
			return stacktrace.Propagate(err, "An error occurred processing bulk command #%v of type '%v'", idx, command.Type)
		}
	}
	return nil
}

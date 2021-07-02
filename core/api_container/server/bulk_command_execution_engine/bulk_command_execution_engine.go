/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package bulk_command_execution_engine

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/bulk_command_execution"
	"github.com/kurtosis-tech/kurtosis/api_container/server/bulk_command_execution_engine/v0_bulk_command_execution"
	"github.com/palantir/stacktrace"
)

type deserializableBulkCommandsDocument struct {
	bulk_command_execution.VersionedBulkCommandsDocument
	BodyBytes     json.RawMessage 						`json:"body"`
}

type BulkCommandExecutionEngine struct {
	v0BulkCommandsProcessor *v0_bulk_command_execution.V0BulkCommandProcessor
}

func NewBulkCommandExecutionEngine(v0BulkCommandsProcessor *v0_bulk_command_execution.V0BulkCommandProcessor) *BulkCommandExecutionEngine {
	return &BulkCommandExecutionEngine{v0BulkCommandsProcessor: v0BulkCommandsProcessor}
}

func (processor BulkCommandExecutionEngine) Process(ctx context.Context, serializedBulkCommandsDocument []byte) error {
	bulkCommandsDocument := new(deserializableBulkCommandsDocument)
	if err := json.Unmarshal(serializedBulkCommandsDocument, &bulkCommandsDocument); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the serialized bulk command bytes to check the schema version")
	}
	switch bulkCommandsDocument.SchemaVersion {
	case bulk_command_execution.V0:
		if err := processor.v0BulkCommandsProcessor.Process(ctx, bulkCommandsDocument.BodyBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred processing the v%v commands", bulkCommandsDocument.SchemaVersion)
		}
	default:
		return stacktrace.NewError("Unrecognized schema version '%v'", bulkCommandsDocument.SchemaVersion)
	}
	return nil
}

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

package bulk_command_execution

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/bulk_command_execution/v0_bulk_command_api"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	latestSchemaVersion = V0
)

type VersionedBulkCommandsDocument struct {
	SchemaVersion SchemaVersion `json:"schemaVersion"`
}

type serializableBulkCommandsDocument struct {
	VersionedBulkCommandsDocument
	Body	interface{}			`json:"body"`
}

type BulkCommandSerializer struct {}

func NewBulkCommandSerializer() *BulkCommandSerializer {
	return &BulkCommandSerializer{}
}

func (serializer BulkCommandSerializer) Serialize(bulkCommands v0_bulk_command_api.V0BulkCommands) ([]byte, error) {
	toSerialize := serializableBulkCommandsDocument{
		VersionedBulkCommandsDocument: VersionedBulkCommandsDocument{
			SchemaVersion: latestSchemaVersion,
		},
		Body:                          bulkCommands,
	}
	bytes, err := json.Marshal(toSerialize)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing bulk commands to bytes")
	}
	return bytes, nil
}



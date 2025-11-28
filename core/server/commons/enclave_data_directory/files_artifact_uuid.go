/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

type FilesArtifactUUID string

func NewFilesArtifactUUID() (FilesArtifactUUID, error) {
	newIDStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not generate Universally Unique ID.")
	}
	newFilesArtifactUuid := FilesArtifactUUID(newIDStr)
	return newFilesArtifactUuid, nil
}

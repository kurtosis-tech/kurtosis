/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/google/uuid"
	"github.com/kurtosis-tech/stacktrace"
)

type FilesArtifactID string

func NewFilesArtifactID() (FilesArtifactID, error) {
	newIDStr, err := getUniversallyUniqueID()
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not generate Universally Unique ID.")
	}
	newFilesArtifactId := FilesArtifactID(newIDStr)
	return newFilesArtifactId, nil
}

// There are some suggestions that go's implementation of uuid is not RFC compliant.
// If we can verify it is compliant, it would be better to use ipv6 as nodeID and interface name where the data came in.
// Just generating a random one for now.
func getUniversallyUniqueID() (string, error) {
	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating a UUID")
	}
	return generatedUUID.String(), nil
}

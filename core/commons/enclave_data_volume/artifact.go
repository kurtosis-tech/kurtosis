/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

type Artifact struct {
	urlHash string
	file *EnclaveDataVolFile
}

func newArtifact(urlHash string, file *EnclaveDataVolFile) *Artifact {
	return &Artifact{urlHash: urlHash, file: file}
}

func (artifact Artifact) GetUrlHash() string {
	return artifact.urlHash
}

func (artifact Artifact) GetFile() *EnclaveDataVolFile {
	return artifact.file
}

/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

// Represents a file inside the enclave data volume
type EnclaveDataVolFile struct {
	absoluteFilepath          string
	filepathRelativeToVolRoot string
}

func newEnclaveDataVolFile(absoluteFilepath string, filepathRelativeToVolRoot string) *EnclaveDataVolFile {
	return &EnclaveDataVolFile{absoluteFilepath: absoluteFilepath, filepathRelativeToVolRoot: filepathRelativeToVolRoot}
}

// Gets the absolute path to the file
func (file EnclaveDataVolFile) GetAbsoluteFilepath() string {
	return file.absoluteFilepath
}

// Gets the path to the file relative to the root of the enclave data volume
func (file EnclaveDataVolFile) GetFilepathRelativeToVolRoot() string {
	return file.filepathRelativeToVolRoot
}



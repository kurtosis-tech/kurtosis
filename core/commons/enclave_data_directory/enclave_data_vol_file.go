/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

// Represents a file inside the enclave data directory
type EnclaveDataDirFile struct {
	absoluteFilepath              string
	filepathRelativeToDataDirRoot string
}

func newEnclaveDataDirFile(absoluteFilepath string, filepathRelativeToDataDirRoot string) *EnclaveDataDirFile {
	return &EnclaveDataDirFile{absoluteFilepath: absoluteFilepath, filepathRelativeToDataDirRoot: filepathRelativeToDataDirRoot}
}

// Gets the absolute path to the file
func (file EnclaveDataDirFile) GetAbsoluteFilepath() string {
	return file.absoluteFilepath
}

// Gets the path to the file relative to the root of the enclave data dir
func (file EnclaveDataDirFile) GetFilepathRelativeToDataDirRoot() string {
	return file.filepathRelativeToDataDirRoot
}



/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"sync"
	"os"
)

/*
Needs:
File Info:
	unique ID to reference a file.
	the location of the file.
	relative location, similar to the file cache system.

File store manager/handler

File Actions:
	Save file to the disk.
		While saving, need to lock up the space set aside for it.
		Ensure the unique ID assigned to the file does not already exist.
	File constructor
*/


type FileStore struct {
	uuid							string
	absolutePath 					string
	dirpathRelativeToDataDirRoot 	string
	mutex 							*sync.Mutex
}

func newFileStore(uuid string, absolutePath string, dirpathRelativeToDaraDirRoot string) *FileStore {
	return &FileStore {
		uuid: 							uuid,
		absolutePath: 					absolutePath,
		dirpathRelativeToDataDirRoot: 	dirpathRelativeToDaraDirRoot,
		mutex: 							&sync.Mutex{},
	}
}

func AddFile(file *os.File) {
	/*
	Needs to:
		lock the data sector for writing.
		Create a uuid.
		Create a path for the file to be stored.
		Save the path and its relative path to root as a string
		Create the file store object and fill with information
	*/
}

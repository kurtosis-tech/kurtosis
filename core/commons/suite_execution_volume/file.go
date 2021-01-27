/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

// Represents a file inside the suite execution volume
type File struct {
	absoluteFilepath          string
	filepathRelativeToVolRoot string
}

func newFile(absoluteFilepath string, filepathRelativeToVolRoot string) *File {
	return &File{absoluteFilepath: absoluteFilepath, filepathRelativeToVolRoot: filepathRelativeToVolRoot}
}



// Gets the absolute path to the file
func (file File) GetAbsoluteFilepath() string {
	return file.absoluteFilepath
}

// Gets the path to the file relative to the root of the suite execution volume
func (file File) GetFilepathRelativeToVolRoot() string {
	return file.filepathRelativeToVolRoot
}



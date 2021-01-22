/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

// Represents a file inside the suite execution volume
type File struct {
	absoluteFilepath string
	relativeFilepath string
}

func newFile(absoluteFilepath string, relativeFilepath string) *File {
	return &File{absoluteFilepath: absoluteFilepath, relativeFilepath: relativeFilepath}
}

// Gets the absolute path to the file inside the mounted suite execution volume
func (file File) GetAbsoluteFilepath() string {
	return file.absoluteFilepath
}

// Gets the path to the file relative to the root of the suite execution volume
func (file File) GetRelativeFilepath() string {
	return file.relativeFilepath
}



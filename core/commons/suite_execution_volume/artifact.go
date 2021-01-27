/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

type Artifact struct {
	urlHash string
	file *File
}

func newArtifact(urlHash string, file *File) *Artifact {
	return &Artifact{urlHash: urlHash, file: file}
}

func (artifact Artifact) GetUrlHash() string {
	return artifact.urlHash
}

func (artifact Artifact) GetFile() *File {
	return artifact.file
}

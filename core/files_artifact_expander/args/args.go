/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
)

// JSON-serialized args that the files artifact expander container take in
const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type FilesArtifactExpanderArgs struct {

	APIContainerIpAddress string `json:"apiContainerIpAddress"`
	ApiContainerPort        uint16                   `json:"apiContainerPort"`
	FilesArtifactExpansions []FilesArtifactExpansion `json:"filesArtifactExpansions"`

}

type FilesArtifactExpansion struct {
	FilesArtifactId string `json:"filesArtifactId"`
	DirPathToExpandTo string `json:"dirPathToExpandTo"`
}

func (args *FilesArtifactExpanderArgs) UnmarshalJSON(data []byte) error {
	type FilesArtifactExpanderArgsMirror FilesArtifactExpanderArgs
	var filesArtifactExpanderArgsMirror FilesArtifactExpanderArgsMirror
	if err := json.Unmarshal(data, &filesArtifactExpanderArgsMirror); err != nil {
		return stacktrace.Propagate(err, "Failed to unmarhsal files artifact expander args")
	}
	*args = FilesArtifactExpanderArgs(filesArtifactExpanderArgsMirror)
	return nil
}

func NewFilesArtifactExpanderArgs(apiContainerIpAddress string, apiContainerPort uint16, filesArtifactExpansions []FilesArtifactExpansion) *FilesArtifactExpanderArgs {
	result := &FilesArtifactExpanderArgs{
		APIContainerIpAddress:   apiContainerIpAddress,
		ApiContainerPort:        apiContainerPort,
		FilesArtifactExpansions: filesArtifactExpansions,
	}

	return result
}
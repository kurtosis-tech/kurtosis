/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
	"strings"
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

func NewFilesArtifactExpanderArgs(apiContainerIpAddress string, apiContainerPort uint16, filesArtifactExpansions []FilesArtifactExpansion) (*FilesArtifactExpanderArgs, error) {
	result := &FilesArtifactExpanderArgs{
		APIContainerIpAddress:   apiContainerIpAddress,
		ApiContainerPort:        apiContainerPort,
		FilesArtifactExpansions: filesArtifactExpansions,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "Expected args object to be valid, instead an error occurred validating it")
	}

	return result, nil
}

func (args *FilesArtifactExpanderArgs) validate() error {
	// Generic validation based on field type
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i)
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}
	return nil
}
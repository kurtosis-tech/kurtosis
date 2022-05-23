/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"github.com/kurtosis-tech/stacktrace"
	"reflect"
	"strings"
)

// JSON-serialized args that the files artifacts expander container take in
const (
	jsonFieldTag = "json"
)

// Fields are public for JSON de/serialization
type FilesArtifactsExpanderArgs struct {
	APIContainerIpAddress string `json:"apiContainerIpAddress"`
	ApiContainerPort        uint16                   `json:"apiContainerPort"`
	FilesArtifactExpansions []FilesArtifactExpansion `json:"filesArtifactExpansions"`

}

type FilesArtifactExpansion struct {
	FilesArtifactId string `json:"filesArtifactId"`
	DirPathToExpandTo string `json:"dirPathToExpandTo"`
}

func NewFilesArtifactExpanderArgs(apiContainerIpAddress string, apiContainerPort uint16, filesArtifactExpansions []FilesArtifactExpansion) (*FilesArtifactsExpanderArgs, error) {
	result := &FilesArtifactsExpanderArgs{
		APIContainerIpAddress:   apiContainerIpAddress,
		ApiContainerPort:        apiContainerPort,
		FilesArtifactExpansions: filesArtifactExpansions,
	}
	if err := result.validate(); err != nil {
		return nil, stacktrace.Propagate(err, "Expected args object to be valid, instead an error occurred validating it")
	}

	return result, nil
}

func (args *FilesArtifactsExpanderArgs) validate() error {
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
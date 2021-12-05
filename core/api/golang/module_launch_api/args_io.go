/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launch_api

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"reflect"
	"strings"
)

const (
	// All module containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the module container ought to behave
	serializedArgsEnvVar = "SERIALIZED_ARGS"

	jsonFieldTag          = "json"
)

// Intended to be used when starting the container - gets the environment variables that the container should be started with
func GetEnvFromArgs(args *ModuleContainerArgs) (map[string]string, error) {
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing module container args to JSON")
	}

	argsStr := string(argsBytes)
	return map[string]string{
		serializedArgsEnvVar: argsStr,
	}, nil
}

// Intended to be used in the container main.go function - gets args from the environment
func GetArgsFromEnv() (*ModuleContainerArgs, error) {
	serializedParamsStr, found := os.LookupEnv(serializedArgsEnvVar)
	if !found {
		return nil, stacktrace.NewError("No serialized args variable '%v' defined", serializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, stacktrace.NewError("Found serialized args environment variable '%v', but the value was empty", serializedArgsEnvVar)
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args ModuleContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", serializedParamsStr)
	}

	// Generic validation based on field type
	reflectVal := reflect.ValueOf(args)
	reflectValType := reflectVal.Type()
	for i := 0; i < reflectValType.NumField(); i++ {
		field := reflectValType.Field(i);
		jsonFieldName := field.Tag.Get(jsonFieldTag)

		// Ensure no empty strings
		strVal := reflectVal.Field(i).String()
		if strings.TrimSpace(strVal) == "" {
			return nil, stacktrace.NewError("JSON field '%s' is whitespace or empty string", jsonFieldName)
		}
	}

	return &args, nil
}

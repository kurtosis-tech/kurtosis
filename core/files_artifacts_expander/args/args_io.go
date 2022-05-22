/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

// Serialize args to JSON
// Deserialize args from SERIALIZED_PARAMS variable
const (
	// All files artifacts expanders accept exactly one environment variable, which contains the serialized params that
	// dictate how the files artifacts expander ought to behave
	serializedArgsEnvVar = "SERIALIZED_ARGS"
)

func GetEnvFromArgs(args *FilesArtifactsExpanderArgs) (resultEnvVars map[string]string, resultErr error) {
	argBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to to serialize files artifacts expander args to JSON, instead a non nil error was returned")
	}
	argsStr := string(argBytes)

	envVars := map[string]string {
		serializedArgsEnvVar: argsStr,
	}
	return envVars, nil
}

func GetArgsFromEnv() (*FilesArtifactsExpanderArgs, error) {
	serializedParamsStr, found := os.LookupEnv(serializedArgsEnvVar)
	if !found {
		return nil, stacktrace.NewError("Expected to find args environment variable '%v', instead found no such environment variable", serializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, stacktrace.NewError("Expected serialized args environment variable to not be empty, instead it was empty")
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args FilesArtifactsExpanderArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to deserialize args JSON '%v', instead a non-nil error was returned", serializedParamsStr)
	}

	return &args, nil
}

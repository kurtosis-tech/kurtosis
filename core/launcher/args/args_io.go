package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	// All API containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the API container ought to behave
	serializedArgsEnvVar = "SERIALIZED_ARGS"
)

// Intended to be used when starting the container - gets the environment variables that the container should be started with
func GetEnvFromArgs(args *APIContainerArgs) (map[string]string, error) {
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing API container args to JSON")
	}

	argsStr := string(argsBytes)

	return map[string]string{
		serializedArgsEnvVar: argsStr,
	}, nil
}

// Intended to be used in the container main.go function - gets args from the environment
func GetArgsFromEnv() (*APIContainerArgs, error) {
	// TODO TODO TODO MAKE SURE POLYMORPHIC DESERIALIZATION WORKS FOR K8S VS DOCKER BACKENDS
	serializedParamsStr, found := os.LookupEnv(serializedArgsEnvVar)
	if !found {
		return nil, stacktrace.NewError("No serialized args variable '%v' defined", serializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, stacktrace.NewError("Found serialized args environment variable '%v', but the value was empty", serializedArgsEnvVar)
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args APIContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", serializedParamsStr)
	}
	return &args, nil
}
package api_container_launcher

import (
	"encoding/json"
	"github.com/palantir/stacktrace"
	"os"
)

func RetrieveAPIContainerArgs() (*APIContainerArgs, error) {
	serializedParamsStr, found := os.LookupEnv(serializedArgsEnvVar)
	if !found {
		return nil, stacktrace.NewError("No serialized args variable '%v' defined", serializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, stacktrace.NewError("No serialized params were provided")
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args APIContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", serializedParamsStr)
	}
	return &args, nil
}

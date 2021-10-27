package v0

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib/api_container_docker_consts"
	"github.com/palantir/stacktrace"
	"os"
)

func RetrieveAPIContainerArgs() (*api_container_launcher.APIContainerArgs, error) {
	serializedParamsStr, found := os.LookupEnv(api_container_docker_consts.SerializedArgsEnvVar)
	if !found {
		return nil, stacktrace.NewError("No serialized args variable '%v' defined", api_container_docker_consts.SerializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, stacktrace.NewError("No serialized params were provided")
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args api_container_launcher.APIContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", serializedParamsStr)
	}
	return &args, nil
}

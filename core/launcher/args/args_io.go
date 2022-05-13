package args

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"os"
)

const (
	// All API containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the API container ought to behave
	serializedArgsEnvVar = "SERIALIZED_ARGS"

	// The API container needs to know its own IP address so that it can instruct modules how to connect to it
	// The KurtosisBackend can populate environment variables with own IP addrsses, and this is the environment
	// variable that we'll ask the backend to use
	ownIpAddressEnvVar = "OWN_IP_ADDRESS"
)

// Intended to be used when starting the container - gets the environment variables that the container should be started with
func GetEnvFromArgs(args *APIContainerArgs) (
	resultEnvVars map[string]string,
	resultOwnIpAddressEnvVar string,
	resultErr error,
) {
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred serializing API container args to JSON")
	}

	argsStr := string(argsBytes)

	envVars := map[string]string{
		serializedArgsEnvVar: argsStr,
	}
	return envVars, ownIpAddressEnvVar, nil
}

// Intended to be used in the container main.go function - gets args + own IP from the environment variables
func GetArgsFromEnv() (*APIContainerArgs, net.IP, error) {
	serializedParamsStr, found := os.LookupEnv(serializedArgsEnvVar)
	if !found {
		return nil, nil, stacktrace.NewError("No serialized args environment variable '%v' defined", serializedArgsEnvVar)
	}
	if serializedParamsStr == "" {
		return nil, nil, stacktrace.NewError("Found serialized args environment variable '%v', but the value was empty", serializedArgsEnvVar)
	}
	paramsJsonBytes := []byte(serializedParamsStr)
	var args APIContainerArgs
	if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred deserializing the args JSON '%v'", serializedParamsStr)
	}

	ownIpAddrStr, found := os.LookupEnv(ownIpAddressEnvVar)
	if !found {
		return nil, nil, stacktrace.NewError("No own IP address environment variable '%v' defined", ownIpAddressEnvVar)
	}
	if ownIpAddrStr == "" {
		return nil, nil, stacktrace.NewError("Found own IP address environment variable '%v', but the value was empty", ownIpAddrStr)
	}
	ownIpAddr := net.ParseIP(ownIpAddrStr)
	if ownIpAddr == nil {
		return nil, nil, stacktrace.NewError("An error occurred parsing own IP address string '%v'", ownIpAddrStr)
	}

	return &args, ownIpAddr, nil
}
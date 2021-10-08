package api_container_docker_consts

const (
	// All API containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the API container ought to behave
	SerializedArgsEnvVar = "SERIALIZED_ARGS"

	// The location where the Docker volume for storing enclave data will be mounted on the API container
	// This COULD possibly vary across launcher API verisons, but we can deal with that bridge when we come to it
	EnclaveDataVolumeMountpoint = "/kurtosis-enclave-data"
)

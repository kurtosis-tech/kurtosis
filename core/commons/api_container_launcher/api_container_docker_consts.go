package api_container_launcher

const (
	// All API containers accept exactly one environment variable, which contains the serialized params that
	// dictate how the API container ought to behave
	SerializedArgsEnvVar = "SERIALIZED_ARGS"

	// The location where the directory on the Docker host machine for storing enclave data will be bind-mounted
	//  on the API container
	// This COULD possibly vary across launcher API verisons, but we can cross that bridge when we come to it
	EnclaveDataDirMountpoint = "/kurtosis-enclave-data"
)

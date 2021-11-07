package kurtosis_engine_server_docker_api

const (
	// Environment variable containing JSON-serialized args to the engine container
	SerializedArgsEnvVar = "SERIALIZED_ARGS"

	// TODO When the API is inside of this container, make this field private and instead expose a public EngineLauncher
	//  object for starting engine containers
	// This is the directory where the engine will write its data
	// This directory is expected to be bind-mounted from the host machine
	EngineDataDirpathOnEngineContainer = "/engine-data"
)

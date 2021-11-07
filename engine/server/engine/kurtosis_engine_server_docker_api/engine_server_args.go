package kurtosis_engine_server_docker_api

type EngineServerArgs struct {
	LogLevelStr string	`json:"logLevelStr"`

	// The engine needs to know about this so it knows what filepath on the host machine to use when bind-mounting
	//  enclave data directories to the API container & services that the APIC starts
	EngineDataDirpathOnHostMachine string	`json:"engineDataDirpathOnHostMachine"`
}

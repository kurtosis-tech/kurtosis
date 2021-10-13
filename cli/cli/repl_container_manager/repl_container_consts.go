package repl_container_manager

const (
	enclaveDataVolMountpointOnReplContainer = "/kurtosis-enclave-data"

	// This is the directory in which the node REPL is running inside the REPL container, which is where
	//  we'll bind-mount the host machine's current directory into the container so the user can access
	//  files on their host machine
	workingDirpathInsideReplContainer = "/repl"

	replContainerSuccessExitCode = 0

	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
	// vvvvvvvvvvvvvvv If you change these, update the REPL Dockerfile!!! vvvvvvvvvvvv
	replContainerKurtosisSocketEnvVar = "KURTOSIS_API_SOCKET"
	replContainerEnclaveDataVolMountpointEnvVar = "ENCLAVE_DATA_VOLUME_MOUNTPOINT"
	// ^^^^^^^^^^^^^^^ If you change these, update the REPL Dockerfile!!! ^^^^^^^^^^^^
	// WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING WARNING
)

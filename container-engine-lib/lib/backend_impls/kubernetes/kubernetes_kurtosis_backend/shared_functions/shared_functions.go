package shared_functions

import "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"

// TODO Remove this once we split apart the KubernetesKurtosisBackend into multiple backends (which we can only
//  do once the CLI no longer makes any calls directly to the KurtosisBackend, and instead makes all its calls through
//  the API container & engine APIs)
type CliModeArgs struct {
	// No CLI mode args needed for now
}

type ApiContainerModeArgs struct {
	OwnEnclaveId enclave.EnclaveID

	OwnNamespaceName string

	storageClassName string

	// TODO make this more dynamic - maybe guess based on the files artifact size?
	filesArtifactExpansionVolumeSizeInMegabytes uint
}

type EngineServerModeArgs struct {
	/*
		StorageClass name to be used for volumes in the cluster
		StorageClasses must be defined by a cluster administrator.
		passes this in when starting Kurtosis with Kubernetes.
	*/
	storageClassName string

	/*
		Enclave availability must be set and defined by a cluster administrator.
		The user passes this in when starting Kurtosis with Kubernetes.
	*/
	enclaveDataVolumeSizeInMegabytes uint
}
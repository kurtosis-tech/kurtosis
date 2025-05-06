package snapshots

func GetMainScriptToExecuteFromSnapshotPackage(packageRootPathOnDisk string) (string, error) {

	// - create the persistent directory docker managed volumes

	// - service network registers all the services beforehand so they get started with correct ips?

	// - create a starlark script that
	// 	- upload the files artifacts into the enclave

	// 	- recreates the service configs of all the snapshotted services
	// 		- configs use snapshotted images
	// 		- same entrypoint and cmd as original services
	// 		- same env vars and ports as original service
	// 		- only mounts persistent directories

	// 	- adds services in a way that respects service startup order
	// 		- uses the same service registration if possible
	// 		- uses add_services to parallelize services that can be parallelized
	//
	return "", nil
}

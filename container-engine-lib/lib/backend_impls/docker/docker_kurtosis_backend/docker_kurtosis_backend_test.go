package docker_kurtosis_backend

// func TestDockerKurtosisBackend(t *testing.T) {
// 	dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
// 	require.NoError(t, err)
// 	enclaveFreeIpAddrTrackers := map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker{}
// 	productionMode := false

// 	backend := NewDockerKurtosisBackend(dockerManager, enclaveFreeIpAddrTrackers, nil, productionMode)

// 	err = backend.SnapshotEnclave(context.Background(), enclave.EnclaveUUID("tedi"), "./")
// 	require.NoError(t, err)
// }

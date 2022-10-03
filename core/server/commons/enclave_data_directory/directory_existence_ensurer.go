/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

const (
	// NOTE: It's very important that all directories created inside the enclave data directory are created with 0777
	//  permissions, because:
	//  a) the engine data directory (and containing enclave data dirs) are bind-mounted on the Docker host machine
	//  b) the engine container, API container, and pretty much every Docker container (including user service containers), run as 'root'
	//  c) the Docker host machine will not be running as root
	//  d) For the host machine to be able to read & write files inside the engine data directory, it needs to be able
	//      to access the directories inside the engine data directory
	// The longterm fix to this is probably to:
	//  1) make the engine server data a Docker volume
	//  2) have the engine server expose the engine data directory to the Docker host machine via some filesystem-sharing
	//      server, like NFS or CIFS
	// This way, we preserve the host machine's ability to write to services as if they were local on the filesystem, while
	//  actually having the data live inside a Docker volume. This also sets the stage for Kurtosis-as-a-Service (where bind-mount
	//  the engine data dirpath would be impossible).
	enclaveDataSubdirectoryPerms = 0777
)

func ensureDirpathExists(absoluteDirpath string) error {
	if _, statErr := os.Stat(absoluteDirpath); os.IsNotExist(statErr) {
		if mkdirErr := os.Mkdir(absoluteDirpath, enclaveDataSubdirectoryPerms); mkdirErr != nil {
			return stacktrace.Propagate(
				mkdirErr,
				"Directory '%v' in the enclave data dir didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	// This is necessary because the os.Mkdir might not create the directory with the perms that we want due to the umask
	// Chmod is not affected by the umask, so this will guarantee we get a directory with the perms that we want
	if err := os.Chmod(absoluteDirpath, enclaveDataSubdirectoryPerms); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred setting the permissions on directory '%v' to '%v'",
			absoluteDirpath,
			enclaveDataSubdirectoryPerms,
		)
	}
	return nil
}

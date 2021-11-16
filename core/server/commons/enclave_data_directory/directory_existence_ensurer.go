/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"syscall"
)

const (
	// The system umask is a set of bits that are _subtracted_ from the perms when we create a file
	// We really do want a 0777 directory (see comment below), so we have to set this to 0
	umaskForCreatingDirectory = 0

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
		oldMask := syscall.Umask(umaskForCreatingDirectory)
		defer syscall.Umask(oldMask)
		if mkdirErr := os.Mkdir(absoluteDirpath, enclaveDataSubdirectoryPerms); mkdirErr != nil {
			return stacktrace.Propagate(
				mkdirErr,
				"Directory '%v' in the enclave data dir didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	return nil
}

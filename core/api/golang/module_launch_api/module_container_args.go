/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launch_api


type ModuleContainerArgs struct {
	// The port number that the module should listen on
	ListenPortNum uint16	`json:"listenPortNum"`

	// IP:port of the Kurtosis API container
	ApiContainerSocket string	`json:"apiContainerSocket"`

	// Arbitrary serialized data that the module can consume at startup to modify its behaviour
	// Analogous to the "constructor"
	SerializedCustomParams string	`json:"serializedCustomParams"`

	// The location on the module container where the enclave data directory has been mounted during launch
	EnclaveDataDirMountpoint string	`json:"enclaveDataDirMountpoint"`
}

func NewModuleContainerArgs(listenPortNum uint16, apiContainerSocket string, serializedCustomParams string, enclaveDataDirMountpoint string) *ModuleContainerArgs {
	return &ModuleContainerArgs{ListenPortNum: listenPortNum, ApiContainerSocket: apiContainerSocket, SerializedCustomParams: serializedCustomParams, EnclaveDataDirMountpoint: enclaveDataDirMountpoint}
}

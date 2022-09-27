/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_launch_api


type ModuleContainerArgs struct {
	// The ID of the enclave that the module will run inside
	EnclaveID string 	`json:"enclaveId"`

	// The port number that the module should listen on
	ListenPortNum uint16	`json:"listenPortNum"`

	// IP:port of the Kurtosis API container
	ApiContainerSocket string	`json:"apiContainerSocket"`

	// Arbitrary serialized data that the module can consume at startup to modify its behaviour
	// Analogous to the "constructor"
	SerializedCustomParams string	`json:"serializedCustomParams"`
}

func NewModuleContainerArgs(enclaveID string, listenPortNum uint16, apiContainerSocket string, serializedCustomParams string) *ModuleContainerArgs {
	return &ModuleContainerArgs{EnclaveID: enclaveID, ListenPortNum: listenPortNum, ApiContainerSocket: apiContainerSocket, SerializedCustomParams: serializedCustomParams}
}

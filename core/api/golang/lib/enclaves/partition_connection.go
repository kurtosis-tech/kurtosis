/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclaves

import (
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	unblockedPartitionConnectionPacketLossValue = 0
	blockedPartitionConnectionPacketLossValue   = 100
	//Random packet loss is specified in the 'tc' command in percent. The smallest possible non-zero value is: 2^32 = 0.0000000232% More info: https://wiki.linuxfoundation.org/networking/netem
	smallestPossibleNonZeroPacketLossValue      = 0.0000000232
	maxPossiblePacketLossValue                  = 100
)

// PartitionConnection To get an instance of this type, use the NewUnblockedPartitionConnection, NewBlockedPartitionConnection or NewSoftPartitionConnection functions
type PartitionConnection interface {
	getPartitionConnectionInfo() *kurtosis_core_rpc_api_bindings.PartitionConnectionInfo
}

// ====================================================================================================
//                                    	 Implementations
// ====================================================================================================
// A PartitionConnection implementation
type partitionConnection struct {
	packetLossPercentage float32
}

func (connection *partitionConnection) getPartitionConnectionInfo() *kurtosis_core_rpc_api_bindings.PartitionConnectionInfo {
	partitionConnectionInfo := binding_constructors.NewPartitionConnectionInfo(connection.packetLossPercentage)
	return partitionConnectionInfo
}

// NewUnblockedPartitionConnection returns a PartitionConnection indicating that there is not a network partition
func NewUnblockedPartitionConnection() PartitionConnection {
	return &partitionConnection{
		packetLossPercentage: unblockedPartitionConnectionPacketLossValue,
	}
}

// NewBlockedPartitionConnection returns a PartitionConnection indicating that there is a hard partition
func NewBlockedPartitionConnection() PartitionConnection {
	return &partitionConnection{
		packetLossPercentage: blockedPartitionConnectionPacketLossValue,
	}
}

// NewSoftPartitionConnection returns a PartitionConnection indicating that there is a soft partition with x% percentage of packet loss
func NewSoftPartitionConnection(packetLossPercentage float32) (PartitionConnection, error) {

	if err := isValidPacketLossValue(packetLossPercentage); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating packet loss value")
	}

	partConnection := &partitionConnection{
		packetLossPercentage: packetLossPercentage,
	}

	return partConnection, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func isValidPacketLossValue(packetLossPercentage float32) error {

	if packetLossPercentage < smallestPossibleNonZeroPacketLossValue || packetLossPercentage > maxPossiblePacketLossValue {
		return stacktrace.NewError("The packet loss percentage value '%v' is not allowed, it should be >= '%v' and <= '%v'", packetLossPercentage, smallestPossibleNonZeroPacketLossValue, maxPossiblePacketLossValue)
	}

	return nil
}

package partition_topology

import "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/partition_topology_db/partition_connection_overrides"

type PartitionConnection struct {
	packetLoss              PacketLoss
	packetDelayDistribution PacketDelayDistribution
}

var (
	ConnectionAllowed = NewPartitionConnection(ConnectionWithNoPacketLoss, ConnectionWithNoPacketDelay)
	ConnectionBlocked = NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)
)

func NewPartitionConnection(packetLoss PacketLoss, packetDelay PacketDelayDistribution) PartitionConnection {
	return PartitionConnection{
		packetLoss:              packetLoss,
		packetDelayDistribution: packetDelay,
	}
}

func (partitionConnection *PartitionConnection) GetPacketLossPercentage() PacketLoss {
	return partitionConnection.packetLoss
}

func (partitionConnection *PartitionConnection) GetPacketDelay() PacketDelayDistribution {
	return partitionConnection.packetDelayDistribution
}

func newPartitionConnectionFromDbType(currentPartitionConnectionDbType partition_connection_overrides.PartitionConnection) PartitionConnection {
	return NewPartitionConnection(NewPacketLoss(currentPartitionConnectionDbType.PacketLoss), NewNormalPacketDelayDistribution(currentPartitionConnectionDbType.PacketDelayDistribution.AvgDelayMs, currentPartitionConnectionDbType.PacketDelayDistribution.Jitter, currentPartitionConnectionDbType.PacketDelayDistribution.Correlation))
}

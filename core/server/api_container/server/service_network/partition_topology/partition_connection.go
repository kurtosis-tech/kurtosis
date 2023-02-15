package partition_topology

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

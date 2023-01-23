package partition_topology

type PartitionConnection struct {
	packetLoss  PacketLoss
	packetDelay PacketDelay
}

var (
	ConnectionAllowed = NewPartitionConnection(ConnectionWithNoPacketLoss, ConnectionWithNoPacketDelay)
	ConnectionBlocked = NewPartitionConnection(ConnectionWithEntirePacketLoss, ConnectionWithNoPacketDelay)
)

func NewPartitionConnection(packetLoss PacketLoss, packetDelay PacketDelay) PartitionConnection {
	return PartitionConnection{
		packetLoss:  packetLoss,
		packetDelay: packetDelay,
	}
}

func (partitionConnection *PartitionConnection) GetPacketLossPercentage() PacketLoss {
	return partitionConnection.packetLoss
}

func (partitionConnection *PartitionConnection) GetPacketDelay() PacketDelay {
	return partitionConnection.packetDelay
}

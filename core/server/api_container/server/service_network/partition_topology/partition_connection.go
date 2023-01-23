package partition_topology

type PartitionConnection struct {
	packetLossPercentage float32
	packetDelay          PacketDelay
}

var (
	ConnectionAllowed = NewPartitionConnection(0, ConnectionWithNoPacketDelay)
	ConnectionBlocked = NewPartitionConnection(100, ConnectionWithNoPacketDelay)
)

func NewPartitionConnection(packetLossPercentage float32, packetDelay PacketDelay) PartitionConnection {
	return PartitionConnection{
		packetLossPercentage: packetLossPercentage,
		packetDelay:          packetDelay,
	}
}

func (partitionConnection *PartitionConnection) GetPacketLossPercentage() float32 {
	return partitionConnection.packetLossPercentage
}

func (partitionConnection *PartitionConnection) GetPacketDelay() PacketDelay {
	return partitionConnection.packetDelay
}

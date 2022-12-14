package partition_topology

type PartitionConnection struct {
	packetLossPercentage float32
}

var (
	ConnectionAllowed = NewPartitionConnection(0)
	ConnectionBlocked = NewPartitionConnection(100)
)

func NewPartitionConnection(packetLossPercentage float32) PartitionConnection {
	return PartitionConnection{
		packetLossPercentage: packetLossPercentage,
	}
}

func (partitionConnection *PartitionConnection) GetPacketLossPercentage() float32 {
	return partitionConnection.packetLossPercentage
}

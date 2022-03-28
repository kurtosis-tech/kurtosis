package enclave

type NetworkConnection struct {
	packetLossPercentage float32
}

func NewNetworkConnection(packetLossPercentage float32) *NetworkConnection {
	return &NetworkConnection{packetLossPercentage: packetLossPercentage}
}

func (networkConnection *NetworkConnection) GetPacketLossPercentage() float32 {
	return networkConnection.packetLossPercentage
}

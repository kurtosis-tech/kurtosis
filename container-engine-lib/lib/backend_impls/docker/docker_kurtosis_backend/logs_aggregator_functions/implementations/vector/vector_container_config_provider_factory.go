package vector

func createVectorContainerConfigProvider(portNumber uint16) *vectorContainerConfigProvider {
	config := newDefaultVectorConfig(portNumber)
	return newVectorContainerConfigProvider(config)
}

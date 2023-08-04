package vector

func createVectorContainerConfigProvider() *vectorContainerConfigProvider {
	config := newDefaultVectorConfig()
	return newVectorContainerConfigProvider(config)
}

package vector

func createVectorContainerConfigProvider(
	httpPortNumber uint16,
) *vectorContainerConfigProvider {
	return newVectorContainerConfigProvider(httpPortNumber)
}

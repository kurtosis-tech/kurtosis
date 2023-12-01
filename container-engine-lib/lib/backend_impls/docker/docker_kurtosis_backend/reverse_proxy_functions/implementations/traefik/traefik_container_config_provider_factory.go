package traefik

func createTraefikContainerConfigProvider(httpPort uint16, dashboardPort uint16) *traefikContainerConfigProvider {
	config := newDefaultTraefikConfig(httpPort, dashboardPort)
	return newTraefikContainerConfigProvider(config)
}

package traefik

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/reverse_proxy"
)

func createTraefikContainerConfigProvider(httpPort uint16, dashboardPort uint16, networkId string) *traefikContainerConfigProvider {
	config := reverse_proxy.NewDefaultReverseProxyConfig(httpPort, dashboardPort, networkId)
	return newTraefikContainerConfigProvider(config)
}

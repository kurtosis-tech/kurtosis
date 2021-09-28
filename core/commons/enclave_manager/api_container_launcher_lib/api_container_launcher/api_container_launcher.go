package api_container_launcher

import (
	"context"
	"github.com/docker/go-connections/nat"
	"net"
)

type APIContainerLauncher interface {
	// All launcher versions must implement this interface. This allows us to force developers, as much as possible,
	// to handle backwards-incompatibility issues.
	Launch(
		ctx context.Context,
		containerName string,
		enclaveId string,
		networkId string,
		subnetMask string,
		gatewayIpAddr net.IP,
		apiContainerIpAddr net.IP,
		otherTakenIpAddrsInEnclave []net.IP,
		isPartitioningEnabled bool,
		shouldPublishAllPorts bool,
	) (string, *nat.PortBinding, error)
}

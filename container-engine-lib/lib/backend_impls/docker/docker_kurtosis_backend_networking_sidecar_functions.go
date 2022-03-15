package docker

import "context"

func (backendCore *DockerKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	image string,


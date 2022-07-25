package service

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifacts_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
)

// Config options for the underlying container of a service
type ServiceConfig struct {
	containerImageName string

	privatePorts map[string]*port_spec.PortSpec

	publicPorts map[string]*port_spec.PortSpec //TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution

	entrypointArgs []string

	cmdArgs []string

	envVars map[string]string

	// Leave as nil to not do any files artifact expansion
	filesArtifactExpansion *files_artifacts_expansion.FilesArtifactsExpansion

	cpuAllocationMillicpus uint64

	memoryAllocationMegabytes uint64
}
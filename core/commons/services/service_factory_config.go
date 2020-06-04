package services

import "os"

// Contains configuration determining what type of objects the ServiceFactory will produce
// This is implicitly a DockerContainerServiceFactoryConfig; we could abstract it easily if we wanted other foundations for services
type ServiceFactoryConfig interface {
	GetDockerImage() string

	GetUsedPorts() map[int]bool

	// TODO when Go gets generics, make the type of 'dependencies' be the same as the output of GetStartCommand
	// If Go had generics, dependencies should be of type []T
	GetStartCommand(publicIpAddr string, dependencies []Service) []string

	// If Go had generics, the return type would be T
	GetServiceFromIp(ipAddr string) Service

	GetFilepathsToMount() map[string]bool

	InitializeMountedFiles(mountedFiles map[string]*os.File)
}


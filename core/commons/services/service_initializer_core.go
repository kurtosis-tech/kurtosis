package services

import (
	"github.com/docker/go-connections/nat"
	"os"
)

// TODO When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
// Contains configuration determining what type of objects the ServiceInitializer will produce
// This is implicitly a DockerContainerServiceFactoryConfig; we could abstract it easily if we wanted other foundations for services
type ServiceInitializerCore interface {
	GetUsedPorts() map[nat.Port]bool

	// TODO when Go gets generics, make the type of 'dependencies' to be []N
	// If Go had generics, dependencies should be of type []T
	/*
	Builds the command that the Docker image will be run with out of the given arguments
	Args:
		mountedFileFilepaths: Mapping between the file keys returned from GetFilesToMount and the actual filepaths on the Docker image
		publicIpAddr: The IP address of the Docker image running the service
		dependencies: The Services that this service depends on (for use in case the command line to the service changes based on dependencies)
	 */
	GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr string, dependencies []Service) ([]string, error)

	// Return a filepath on the image where the test volume will be mounted
	GetTestVolumeMountpoint() string

	// TODO When Go has generics, make this return type to be S
	GetServiceFromIp(ipAddr string) Service

	/*
	This method should return a set of identifiers such that one file will be created (in a location abstracted from the initializer core) and then passed
	 to the initializer core's InitializeMountedFiles method

	NOTE: The keys returned here are ONLY used for identification purposes in InitializeMountedFiles - they won't be used in the file names or paths!
	 */
	GetFilesToMount() map[string]bool

	InitializeMountedFiles(mountedFiles map[string]*os.File, dependencies []Service) error
}


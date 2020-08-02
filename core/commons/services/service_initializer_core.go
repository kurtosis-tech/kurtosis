package services

import (
	"github.com/docker/go-connections/nat"
	"net"
	"os"
)

// GENERICS TOOD: When Go has generics, parameterize this to be <N, S extends N> where S is the
//  specific service interface and N represents the interface that every node on the network has
/*
Tells Kurtosis how to create a Docker container representing a user-defined service in the test network.
 */
type ServiceInitializerCore interface {
	// Gets the "set" of ports that the Docker container running the service will listen on
	GetUsedPorts() map[nat.Port]bool

	// GENERICS TOOD: When Go has generics, make this return type be parameterized
	/*
	Uses the IP address of the Docker container running the service to create an implementation of the interface the developer
	has created to represent their service.

	NOTE: Because Go doesn't have generics, we can't properly parameterize the return type to be the actual service interface
	that the developer has created; nonetheless, the developer should return an implementation of their interface (which itself
	should extend Service).

	Args:
		ipAddr: The IP address of the Docker container running the service
	*/
	GetServiceFromIp(ipAddr string) Service

	/*
	This method is used to declare that the service will need a set of files in order to run. To do this, the developer
	declares a set of string keys that are meaningful to the developer, and Kurtosis will create one file per key. These newly-createed
	file objects will then be passed in to the `InitializeMountedFiles` and `GetStartCommand` functions below keyed on the
	strings that the developer passed in, so that the developer can initialize the contents of the files as they please.
	Kurtosis then guarantees that these files will be made available to the service at startup time.

	NOTE: The keys that the developer returns here are ONLY used for developer identification purposes; the actual
	filenames and filepaths of the file are implementation details handled by Kurtosis!

	Returns:
		A "set" of user-defined key strings identifying the files that the service will need, which is how files will be
			identified in `InitializeMountedFiles` and `GetStartCommand`
	*/
	GetFilesToMount() map[string]bool

	/*
	Initializes the contents of the files that the developer requested in `GetFilesToMount` with whatever contents the developer desires

	Args:
		mountedFiles: A mapping of developer_key -> file_pointer, with developer_key corresponding to the keys declares in
			`GetFilesToMount`
		dependencies: The services that this service depends on (which, depending on the service, might be necessary
			for filling in config file values)
	 */
	InitializeMountedFiles(mountedFiles map[string]*os.File, dependencies []Service) error

	/*
	Kurtosis mounts the files that the developer requested in `GetFilesToMount` via a Docker volume, but Kurtosis doesn't
	know anything about the Docker image backing the service so therefore doesn't know what filepath it can safely mount
	the volume on. This function uses the developer's knowledge of the Docker image running the service to inform
	Kurtosis of a filepath where the Docker volume can be safely mounted.

	Returns:
		A filepath on the Docker image backing this service that's safe to mount the test volume on
	 */
	GetTestVolumeMountpoint() string

	// GENERICS TOOD: when Go gets generics, make the type of 'dependencies' to be []N
	// If Go had generics, dependencies should be of type []T
	/*
	Uses the given arguments to build the command that the Docker container running this service will be launched with.

	Args:
		mountedFileFilepaths: Mapping of developer_key -> initialized_file_filepath where developer_key corresponds to the keys returned
			in the `GetFilesToMount` function, and initialized_file_filepath is the path *on the Docker container* of where the
			file has been mounted. The files will have already been initialized via the `InitializeMountedFiles` function.
		publicIpAddr: The IP address of the Docker image running the service
		dependencies: The services that this service depends on (for use in case the command line to the service changes based on dependencies)

	Returns:
		The command fragments which will be used to construct the run command which will be used to launch the Docker container
			running the service
	 */
	GetStartCommand(mountedFileFilepaths map[string]string, publicIpAddr net.IP, dependencies []Service) ([]string, error)

}


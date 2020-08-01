package services

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
	"os"
	"path/filepath"
)

/*
A struct that wraps a user-defined ServiceInitializerCore, which will instruct the initializer how to launch a new instance
	of the user's service.
 */
type ServiceInitializer struct {
	// The user-defined instructions for how to initialize their service
	core ServiceInitializerCore

	// The name of the Docker network that the new service should be added to
	networkName string

	// The path to the directory where the test volume is mounted on the CONTROLLER Docker image. We need to know this
	// 	because this is where this initializer will create the files required by the service being initialized.
	testVolumeControllerDirpath string
}

/*
Creates a new service initializer that will initialize services using the user-defined core.

Args:
	core: The user-defined logic for instantiating their particular service
	networkName: The name of the Docker network that the service will be added to
	testVolumeControllerDirpath: The dirpath where the test Docker volume is mounted on the test controller Docker container
 */
func NewServiceInitializer(core ServiceInitializerCore, networkName string, testVolumeControllerDirpath string) *ServiceInitializer {
	return &ServiceInitializer{
		core: core,
		networkName: networkName,
		testVolumeControllerDirpath: testVolumeControllerDirpath,
	}
}

// If Go had generics, this would be genericized so that the arg type = return type
/*
Creates a service with the given parameters

Args:
	context: Context that the creation of the service is running in (used for cancellation)
	testVolumeName: The name of the test Docker volume that will be mounted on the Docker container running the service
	dockerImage: The name of the Docker image that the new service will be started with
	staticIp: The IP the new service will be given
	manager: The DockerManager used to launch the container running the service
	dependencies: The services that the service-to-be-started depends on

Returns:
	Service: The interface which should be used to access the newly-created service (which, because Go doesn't have generics,
		will need to be casted to the appropriate type)
	string: The ID of the Docker container the service is running in
 */
func (initializer ServiceInitializer) CreateService(
			context context.Context,
			testVolumeName string,
			dockerImage string,
			staticIp string,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	initializerCore := initializer.core
	usedPorts := initializerCore.GetUsedPorts()

	serviceDirname := fmt.Sprintf("service-%v", uuid.Generate().String())
	controllerServiceDirpath := filepath.Join(initializer.testVolumeControllerDirpath, serviceDirname)
	err := os.Mkdir(controllerServiceDirpath, os.ModeDir)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "An error occurred creating the new service's directory in the volume at filepath '%v'", controllerServiceDirpath)
	}
	mountServiceDirpath := filepath.Join(initializerCore.GetTestVolumeMountpoint(), serviceDirname)

	requestedFiles := initializerCore.GetFilesToMount()
	osFiles := make(map[string]*os.File)
	mountFilepaths := make(map[string]string)
	for fileId, _ := range requestedFiles {
		filename := uuid.Generate().String()
		hostFilepath := filepath.Join(controllerServiceDirpath, filename)
		fp, err := os.Create(hostFilepath)
		if err != nil {
			return nil, "", stacktrace.Propagate(err, "Could not create new file for requested file ID '%v'", fileId)
		}
		defer fp.Close()
		osFiles[fileId] = fp
		mountFilepaths[fileId] = filepath.Join(mountServiceDirpath, filename)
	}
	err = initializerCore.InitializeMountedFiles(osFiles, dependencies)
	startCmdArgs, err := initializerCore.GetStartCommand(mountFilepaths, staticIp, dependencies)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Failed to create start command.")
	}

	volumeMounts := map[string]string{
		testVolumeName: initializerCore.GetTestVolumeMountpoint(),
	}

	ipAddr, containerId, err := manager.CreateAndStartContainer(
			context,
			dockerImage,
			initializer.networkName,
			staticIp,
			usedPorts,
			startCmdArgs,
			make(map[string]string),
			make(map[string]string),
			volumeMounts)
	if err != nil {
		return nil, "", stacktrace.Propagate(err, "Could not start docker service for image %v", dockerImage)
	}
	return initializer.core.GetServiceFromIp(ipAddr), containerId, nil
}

/*
Calls down to the initializer core to get an instance of the user-defined interface that is used for interacting with
	the user's service. The core will do the instantiation of the actual interface implementation.
 */
func (initializer ServiceInitializer) GetServiceFromIp(ipAddr string) Service {
	return initializer.core.GetServiceFromIp(ipAddr)
}

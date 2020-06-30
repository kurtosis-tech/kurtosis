package services

import (
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
	"os"
	"path/filepath"
)


// This implicitly is a Docker container-backed service initializer, but we could abstract to other backends if we wanted later
type ServiceInitializer struct {
	core ServiceInitializerCore
	networkName string
}

func NewServiceInitializer(core ServiceInitializerCore, networkName string) *ServiceInitializer {
	return &ServiceInitializer{
		core: core,
		networkName: networkName,
	}
}

// If Go had generics, this would be genericized so that the arg type = return type
/*
Creates a service with the given parameters
Args:
	testVolumeName: The name of the test volume to mount on the node
	testVolumeControllerDirpath: The path to the directory where the test volume is mounted on the controller Docker image
	dockerImage: The name of the Docker image that the new service will be started with
	staticIp: The IP the new service will be given
	manager: The DockerManager used to launch the container running the service
	dependencies: The services that the service-to-be-started depends on
 */
func (initializer ServiceInitializer) CreateService(
			testVolumeName string,
			testVolumeControllerDirpath string,
			dockerImage string,
			staticIp string,
			manager *docker.DockerManager,
			dependencies []Service) (Service, string, error) {
	initializerCore := initializer.core
	usedPorts := initializerCore.GetUsedPorts()

	serviceDirname := fmt.Sprintf("service-%v", uuid.Generate().String())
	controllerServiceDirpath := filepath.Join(testVolumeControllerDirpath, serviceDirname)
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

	// TODO we really want GetEnvVariables instead of GetStartCmd because every image should be nicely parameterized to avoid
	//   the testing code knowing about the specifics of the image (like where the binary is located). However, this relies
	//   on the service images being parameterized with environment variables.
	ipAddr, containerId, err := manager.CreateAndStartContainer(
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

func (initializer ServiceInitializer) LoadService(ipAddr string) Service {
	return initializer.core.GetServiceFromIp(ipAddr)
}

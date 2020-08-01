package services

import (
	"context"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/palantir/stacktrace"
	"net"
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
	context: Context that the creation of the service is running in
	testVolumeName: The name of the test volume to mount on the node
	testVolumeControllerDirpath: The path to the directory where the test volume is mounted on the controller Docker image
	dockerImage: The name of the Docker image that the new service will be started with
	staticIp: The IP the new service will be given
	manager: The DockerManager used to launch the container running the service
	dependencies: The services that the service-to-be-started depends on
 */
func (initializer ServiceInitializer) CreateService(
			context context.Context,
			testVolumeName string,
			testVolumeControllerDirpath string,
			dockerImage string,
			staticIp net.IP,
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
	return initializer.core.GetServiceFromIp(ipAddr.String()), containerId, nil
}

func (initializer ServiceInitializer) LoadService(ipAddr net.IP) Service {
	return initializer.core.GetServiceFromIp(ipAddr.String())
}

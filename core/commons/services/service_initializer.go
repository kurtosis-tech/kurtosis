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
}

func NewServiceInitializer(core ServiceInitializerCore) *ServiceInitializer {
	return &ServiceInitializer{
		core: core,
	}
}

// If Go had generics, this would be genericized so that the arg type = return type
/*
Creates a service with the given parameters
Args:
	testVolumeHostDirpath: The path to the directory of the test volume on the code running CreateService, which will be mounted on the service
	testVolumeMountDirpath: The path to the directory where the test volume will be mounted on the Service's Docker image
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

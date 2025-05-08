package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/docker/docker/client"
	api_services "github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	container_engine_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	path_compression "github.com/kurtosis-tech/kurtosis/path-compression"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type Snapshot struct{}

type SnapshotCreator struct {
	serviceNetwork service_network.ServiceNetwork

	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	snapshotBaseDirPath string
}

func NewSnapshotCreator(serviceNetwork service_network.ServiceNetwork, interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore, snapshotBaseDirPath string) *SnapshotCreator {
	return &SnapshotCreator{
		serviceNetwork:               serviceNetwork,
		interpretationTimeValueStore: interpretationTimeValueStore,
		snapshotBaseDirPath:          snapshotBaseDirPath,
	}
}

// Snapshot package layout
// /snapshot
//
//	/persistent-directories
//	   /persistent-key-1
//		   /tar.tgz
//	   /persistent-key-2
//		   /tar.tgz
//	/files-artifacts
//		files-artifacts-name.tar.tgz
//		...
//	/services
//		/service-name
//			service-config.json
//			image.tar
//			service-registration.json
//	 args.json
//	 return args ...
//	 service-startup-order.txt
//	 files-artifacts-names.txt
//	 persistent-directories-names.txt

func (sc *SnapshotCreator) CreateSnapshot(ctx context.Context) (string, error) {
	// create snapshot directory for this api container and snapshot
	snapshotStoreDir := path.Join("/kurtosis-data", "snapshot-store") // TODO: refactor to use enclave data directory
	if err := os.MkdirAll(snapshotStoreDir, 0755); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating snapshot store directory %s", snapshotStoreDir)
	}

	// create a tmp directory that will hold all the image tars until we save them to the snapshot directory and defer removal of it
	snapshotDir := path.Join(snapshotStoreDir, "tmp") // TODO: refactor to use name of enclave
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating tmp directory %s", snapshotDir)
	}
	servicesDir := path.Join(snapshotDir, "services")
	if err := os.MkdirAll(servicesDir, 0755); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating services directory %s", servicesDir)
	}
	filesArtifactsDir := path.Join(snapshotDir, "files-artifacts")
	if err := os.MkdirAll(filesArtifactsDir, 0755); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating files artifacts directory %s", filesArtifactsDir)
	}
	persistentDirectoriesDir := path.Join(snapshotDir, "persistent-directories")
	if err := os.MkdirAll(persistentDirectoriesDir, 0755); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating persistent directories directory %s", persistentDirectoriesDir)
	}
	defer os.RemoveAll(snapshotDir)

	services, err := sc.serviceNetwork.GetServices(ctx)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting services in enclave")
	}
	logrus.Infof("Services in enclave: %v", services)
	logrus.Infof("Number of services in enclave: %v", len(services))

	for _, service := range services {

		serviceDir := path.Join(servicesDir, service.GetRegistration().GetHostname())
		if err := os.Mkdir(serviceDir, 0755); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred creating directory '%s' for service '%s'", servicesDir, service.GetRegistration().GetHostname())
		}

		// output snapshotted image
		err = sc.outputSnapshottedImage(ctx, serviceDir, service)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred outputting snapshotted image for service %s", service.GetRegistration().GetHostname())
		}
		// TODO: defer undo removal of images

		// output service config
		err = sc.outputServiceConfig(serviceDir, service)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred outputting service config for service %s", service.GetRegistration().GetHostname())
		}
		// TODO: defer undo writing service config

		// output files artifacts

		// output persistent directories

		// output service startup order

		// output return args

		// output input args
	}

	// tar everything in the tmp directory and save to snapshot directory
	compressedFilePath, sizeOfCompressedPath, _, err := path_compression.CompressPathToFile(snapshotDir, false)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating tar for snapshot")
	}
	logrus.Infof("Snapshot tar path and size: %s, %d", compressedFilePath, sizeOfCompressedPath)

	return compressedFilePath, nil
}

func (sc *SnapshotCreator) outputSnapshottedImage(ctx context.Context, serviceDir string, service *container_engine_service.Service) error {
	dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating docker manager")
	}

	// HIDE THIS PART BEHIND KURTOSIS BACKEND
	containers, err := dockerManager.GetContainersByLabels(ctx, map[string]string{
		docker_label_key.IDDockerLabelKey.GetString(): service.GetRegistration().GetHostname(),
	}, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels for service %s", service.GetRegistration().GetHostname())
	}
	if len(containers) == 0 {
		return stacktrace.NewError("No containers found for service %s", service.GetRegistration().GetHostname())
	}
	container := containers[0]
	containerId := container.GetId()
	logrus.Infof("Committing container %v", containerId)

	// commit container to image
	imageName := fmt.Sprintf("%v-%v-snapshot-img", service.GetRegistration().GetHostname(), time.Now().Unix())
	err = dockerManager.CommitContainer(ctx, containerId, imageName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred committing container %v", containerId)
	}

	// save image to snapshot directory
	imagePath := path.Join(serviceDir, imageName)
	err = dockerManager.SaveImage(ctx, imageName, imagePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred saving image to file for service %s", service.GetRegistration().GetHostname())
	}

	// remove the image
	// err = dockerManager.RemoveImage(ctx, imageName, true, true)
	// if err != nil {
	// 	return nil, stacktrace.Propagate(err, "An error occurred removing image %s", imageName)
	// }

	// HIDE THBEHIND KURTOSIS BACKEND

	return nil

}

func (sc *SnapshotCreator) outputServiceConfig(serviceDir string, service *container_engine_service.Service) error {
	serviceConfig, err := sc.interpretationTimeValueStore.GetServiceConfig(container_engine_service.ServiceName(service.GetRegistration().GetHostname()))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service config for service %s", service.GetRegistration().GetHostname())
	}

	jsonServiceConfig, err := convertServiceConfigToJsonServiceConfig(serviceConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred converting service config to json for service %s", service.GetRegistration().GetHostname())
	}

	jsonServiceConfigBytes, err := json.Marshal(jsonServiceConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling service config to json for service %s", service.GetRegistration().GetHostname())
	}

	jsonServiceConfigPath := path.Join(serviceDir, "service-config.json")
	err = os.WriteFile(jsonServiceConfigPath, jsonServiceConfigBytes, 0644)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing service config to file for service %s", service.GetRegistration().GetHostname())
	}

	return nil

}

func convertServiceConfigToJsonServiceConfig(serviceConfig *container_engine_service.ServiceConfig) (api_services.ServiceConfig, error) {
	// create ports
	privatePorts := make(map[string]api_services.Port, len(serviceConfig.GetPrivatePorts()))
	for i, port := range serviceConfig.GetPrivatePorts() {
		privatePorts[i] = api_services.Port{
			Number:                   uint32(port.GetNumber()),
			Transport:                int(port.GetTransportProtocol()),
			MaybeApplicationProtocol: *port.GetMaybeApplicationProtocol(),
			Wait:                     "",
		}
	}
	publicPorts := make(map[string]api_services.Port, len(serviceConfig.GetPublicPorts()))
	for i, port := range serviceConfig.GetPublicPorts() {
		publicPorts[i] = api_services.Port{
			Number:                   uint32(port.GetNumber()),
			Transport:                int(port.GetTransportProtocol()),
			MaybeApplicationProtocol: *port.GetMaybeApplicationProtocol(),
			Wait:                     "",
		}
	}

	// create tolerations
	tolerations := make([]api_services.Toleration, len(serviceConfig.GetTolerations()))
	for i, toleration := range serviceConfig.GetTolerations() {
		tolerations[i] = api_services.Toleration{
			Key:               toleration.Key,
			Value:             toleration.Value,
			Operator:          string(toleration.Operator),
			Effect:            string(toleration.Effect),
			TolerationSeconds: int64(*toleration.TolerationSeconds),
		}
	}

	// create user
	// user := &api_services.User{}
	// if serviceConfig.GetUser() != nil {
	// 	user.UID = uint32(serviceConfig.GetUser().GetUID())
	// }

	// create toleration
	isTiniEnabled := serviceConfig.GetTiniEnabled()
	apiServiceConfig := api_services.ServiceConfig{
		Image: serviceConfig.GetContainerImageName(),
		// PrivatePorts:                privatePorts,
		// PublicPorts:                 publicPorts,
		// Files:                       serviceConfig.GetFilesArtifactsExpansion().ServiceDirpathsToArtifactIdentifiers,
		Entrypoint:                  serviceConfig.GetEntrypointArgs(),
		Cmd:                         serviceConfig.GetCmdArgs(),
		EnvVars:                     serviceConfig.GetEnvVars(),
		PrivateIPAddressPlaceholder: serviceConfig.GetPrivateIPAddrPlaceholder(),
		MaxMillicpus:                uint32(serviceConfig.GetCPUAllocationMillicpus()),
		MinMillicpus:                uint32(serviceConfig.GetMinCPUAllocationMillicpus()),
		MaxMemory:                   uint32(serviceConfig.GetMemoryAllocationMegabytes()),
		MinMemory:                   uint32(serviceConfig.GetMinMemoryAllocationMegabytes()),
		// User:                        user,
		// Tolerations:   tolerations,
		Labels:        serviceConfig.GetLabels(),
		NodeSelectors: serviceConfig.GetNodeSelectors(),
		TiniEnabled:   &isTiniEnabled,
	}

	return apiServiceConfig, nil
}

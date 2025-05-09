package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

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

	snapshotNum int
}

func NewSnapshotCreator(serviceNetwork service_network.ServiceNetwork, interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore, snapshotBaseDirPath string) *SnapshotCreator {
	return &SnapshotCreator{
		serviceNetwork:               serviceNetwork,
		interpretationTimeValueStore: interpretationTimeValueStore,
		snapshotBaseDirPath:          snapshotBaseDirPath,
		snapshotNum:                  0,
	}
}

func (sc *SnapshotCreator) CreateSnapshot() (string, error) {
	ctx := context.Background()
	// create a tmp directory that will hold all the image tars until we save them to the snapshot directory and defer removal of it
	snapshotDir := path.Join(sc.snapshotBaseDirPath, fmt.Sprintf("%d", sc.snapshotNum)) // TODO: refactor to use name of enclave
	if err := os.MkdirAll(snapshotDir, snapshotDirPerms); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating tmp directory %s", snapshotDir)
	}
	sc.snapshotNum++

	servicesDir := path.Join(snapshotDir, snapshotServicesDirPath)
	if err := os.MkdirAll(servicesDir, snapshotDirPerms); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating services directory %s", servicesDir)
	}
	filesArtifactsDir := path.Join(snapshotDir, snapshotFilesArtifactsDirPath)
	if err := os.MkdirAll(filesArtifactsDir, snapshotDirPerms); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating files artifacts directory %s", filesArtifactsDir)
	}
	persistentDirectoriesDir := path.Join(snapshotDir, snapshotPersistentDirectoriesDirPath)
	if err := os.MkdirAll(persistentDirectoriesDir, snapshotDirPerms); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating persistent directories directory %s", persistentDirectoriesDir)
	}
	// defer os.RemoveAll(snapshotDir) // should we remove or keep the snapshot around?

	// TODO: how do you reconcile service network get services
	services, err := sc.serviceNetwork.GetServices(ctx)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting services in enclave")
	}
	logrus.Infof("Services in enclave: %v", services)
	logrus.Infof("Number of services in enclave: %v", len(services))

	serviceNames := []string{}
	for _, service := range services {
		serviceName := service.GetRegistration().GetHostname()
		serviceNames = append(serviceNames, serviceName)
		serviceDir := path.Join(servicesDir, serviceName)
		if err := os.Mkdir(serviceDir, snapshotDirPerms); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred creating directory '%s' for service '%s'", servicesDir, serviceName)
		}

		err = sc.outputSnapshottedImage(ctx, serviceName, serviceDir)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred outputting snapshotted image for service %s", serviceName)
		}
		// TODO: defer undo removal of images

		err = sc.outputServiceConfig(serviceName, serviceDir)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred outputting service config for service %s", serviceName)
		}
		// TODO: defer undo writing service config

		// output service startup order

		// output return args

		// output input args

		// output persistent directories

		// output files artifacts
	}

	// TODO: get the service startup order from the persisted enclave plan
	serviceStartupOrderPath := path.Join(snapshotDir, serviceStartupOrderFileName)
	err = os.WriteFile(serviceStartupOrderPath, []byte(strings.Join(serviceNames, "\n")), snapshotDirPerms)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing service startup order to file '%s'", serviceStartupOrderPath)
	}
	logrus.Infof("Service startup order written to file '%s'", serviceStartupOrderPath)

	// tar everything in the tmp directory and save to snapshot directory
	compressedFilePath, sizeOfCompressedPath, _, err := path_compression.CompressPathToFile(snapshotDir, false)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating tar for snapshot")
	}
	logrus.Infof("Snapshot tar path and size: %s, %d", compressedFilePath, sizeOfCompressedPath)

	return compressedFilePath, nil
}

func (sc *SnapshotCreator) outputSnapshottedImage(ctx context.Context, serviceName, serviceDir string) error {
	dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating docker manager")
	}

	// HIDE THIS PART BEHIND KURTOSIS BACKEND
	containers, err := dockerManager.GetContainersByLabels(ctx, map[string]string{
		docker_label_key.IDDockerLabelKey.GetString(): serviceName,
	}, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels for service %s", serviceName)
	}
	if len(containers) == 0 {
		return stacktrace.NewError("No containers found for service %s", serviceName)
	}
	container := containers[0]
	containerId := container.GetId()
	logrus.Infof("Committing container %v", containerId)

	// commit container to image
	imageName := fmt.Sprintf(snapshottedImageNameFmtSpecifier, serviceName)
	err = dockerManager.CommitContainer(ctx, containerId, imageName)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred committing container %v", containerId)
	}

	// save image to snapshot directory
	imagePath := path.Join(serviceDir, imageName)
	err = dockerManager.SaveImage(ctx, imageName, imagePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred saving image to file for service %s", serviceName)
	}

	// remove the image
	// err = dockerManager.RemoveImage(ctx, imageName, true, true)
	// if err != nil {
	// 	return nil, stacktrace.Propagate(err, "An error occurred removing image %s", imageName)
	// }

	// HIDE THBEHIND KURTOSIS BACKEND

	return nil

}

func (sc *SnapshotCreator) outputServiceConfig(serviceName string, serviceDir string) error {
	serviceConfig, err := sc.interpretationTimeValueStore.GetServiceConfig(container_engine_service.ServiceName(serviceName))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service config for service %s", serviceName)
	}

	jsonServiceConfig, err := convertServiceConfigToJsonServiceConfig(serviceConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred converting service config to json for service %s", serviceName)
	}

	jsonServiceConfigBytes, err := json.Marshal(jsonServiceConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling service config to json for service %s", serviceName)
	}

	jsonServiceConfigPath := path.Join(serviceDir, serviceConfigFileName)
	err = os.WriteFile(jsonServiceConfigPath, jsonServiceConfigBytes, snapshotDirPerms)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing service config to file for service %s", serviceName)
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
			MaybeApplicationProtocol: "http",
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
		Image:        serviceConfig.GetContainerImageName(),
		PrivatePorts: privatePorts,
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

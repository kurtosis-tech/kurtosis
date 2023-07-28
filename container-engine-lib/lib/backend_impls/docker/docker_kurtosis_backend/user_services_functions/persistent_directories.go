package user_service_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

func getOrCreatePersistentDirectories(
	ctx context.Context,
	serviceUuid service.ServiceUUID,
	objAttrsProvider object_attributes_provider.DockerEnclaveObjectAttributesProvider,
	serviceMountpointsToPersistentKey map[string]service_directory.DirectoryPersistentKey,
	dockerManager *docker_manager.DockerManager,
) (map[string]string, error) {
	shouldDeleteVolumes := true
	result := map[string]string{}

	for serviceDirpath, persistentKey := range serviceMountpointsToPersistentKey {
		volumeAttrs, err := objAttrsProvider.ForSinglePersistentDirectoryVolume(serviceUuid, persistentKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error creating persistent directory labels for '%s'", persistentKey)
		}

		volumeName := volumeAttrs.GetName().GetString()
		volumeLabelsStrs := map[string]string{}
		for key, value := range volumeAttrs.GetLabels() {
			volumeLabelsStrs[key.GetString()] = value.GetString()
		}

		potentiallyExistingVolumes, err := dockerManager.GetVolumesByName(ctx, volumeName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred checking for persistent volume existence")
		}
		if len(potentiallyExistingVolumes) == 1 {
			result[serviceDirpath] = volumeName
			continue
		} else if len(potentiallyExistingVolumes) > 1 {
			return nil, stacktrace.NewError("More than one volume with name '%s' exists in docker. This is unexpected", volumeName)
		}

		if err = dockerManager.CreateVolume(ctx, volumeName, volumeLabelsStrs); err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating persistent directory volume '%s' for service '%v'",
				persistentKey,
				serviceUuid,
			)
		}
		result[volumeName] = serviceDirpath
	}

	defer func() {
		if !shouldDeleteVolumes {
			return
		}
		for _, volumeNameStr := range result {
			// Background context so we still run this even if the input context was cancelled
			if err := dockerManager.RemoveVolume(context.Background(), volumeNameStr); err != nil {
				logrus.Warnf(
					"Creating persistent directory volumes didn't complete successfully so we tried to delete volume '%v' that we created, but doing so threw an error:\n%v",
					volumeNameStr,
					err,
				)
				logrus.Warnf("You'll need to clean up volume '%v' manually!", volumeNameStr)
			}
		}
	}()
	shouldDeleteVolumes = false
	return result, nil
}

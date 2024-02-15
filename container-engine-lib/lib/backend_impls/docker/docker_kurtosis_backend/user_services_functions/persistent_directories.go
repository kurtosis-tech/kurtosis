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
	serviceMountpointsToPersistentKey map[string]service_directory.PersistentDirectory,
	dockerManager *docker_manager.DockerManager,
) (map[string]string, error) {
	shouldDeleteVolumes := true
	volumeNamesToRemoveIfFailure := map[string]bool{}
	persistentDirectories := map[string]string{}

	for serviceDirPath, persistentDirectory := range serviceMountpointsToPersistentKey {
		volumeAttrs, err := objAttrsProvider.ForSinglePersistentDirectoryVolume(persistentDirectory.PersistentKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Error creating persistent directory labels for '%s'", persistentDirectory.PersistentKey)
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
			// volume already existed for this enclave, return it and continue to the next one
			persistentDirectories[volumeName] = serviceDirPath
			continue
		} else if len(potentiallyExistingVolumes) > 1 {
			return nil, stacktrace.NewError("More than one volume with name '%s' exists in docker. This is unexpected", volumeName)
		}

		// note this doesn't consider persistentDirectory.Size
		// Docker doesn't support sized volumes - the best you can do is create tmpfs (in memory) persistent volumes
		// but that would go away if you restart Docker
		// TODO Make the `lsp` aware of `size`
		if err = dockerManager.CreateVolume(ctx, volumeName, volumeLabelsStrs); err != nil {
			return nil, stacktrace.Propagate(
				err,
				"An error occurred creating persistent directory volume '%s' for service '%v'",
				persistentDirectory.PersistentKey,
				serviceUuid,
			)
		}
		volumeNamesToRemoveIfFailure[volumeName] = true
		persistentDirectories[volumeName] = serviceDirPath
	}

	defer func() {
		if !shouldDeleteVolumes {
			return
		}
		for volumeNameStr := range volumeNamesToRemoveIfFailure {
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
	return persistentDirectories, nil
}

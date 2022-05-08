/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/current_time_str_provider"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	guidElementSeparator = "-"
)

/*
Class responsible for taking an artifact containing compressed files and creating a new
	files artifact expansion volume and a new the files artifact expander for it.
	The new files artifact expander will be in charge of uncompressing the artifact contents
	into the new volume, and this volume will be mounted on a new service
*/
type FilesArtifactExpander struct {
	kurtosisBackend backend_interface.KurtosisBackend

	enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider

	enclaveId enclave.EnclaveID

	freeIpAddrTracker *lib.FreeIpAddrTracker

	filesArtifactStore *enclave_data_directory.FilesArtifactStore
}

func NewFilesArtifactExpander(kurtosisBackend backend_interface.KurtosisBackend, enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider, enclaveId enclave.EnclaveID, freeIpAddrTracker *lib.FreeIpAddrTracker, filesArtifactStore *enclave_data_directory.FilesArtifactStore) *FilesArtifactExpander {
	return &FilesArtifactExpander{kurtosisBackend: kurtosisBackend, enclaveObjAttrsProvider: enclaveObjAttrsProvider, enclaveId: enclaveId, freeIpAddrTracker: freeIpAddrTracker, filesArtifactStore: filesArtifactStore}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
	ctx context.Context,
	serviceGUID service.ServiceGUID, // Service GUID for whom the artifacts are being expanded into volumes
	artifactUuidsToExpand map[service.FilesArtifactID]bool,
) (map[service.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, error) {

	// TODO PERF: parallelize this to increase speed
	artifactIdsToVolNames := map[service.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName{}
	volumesToDestroyIfSomethingFails := map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool{}
	for filesArtifactId := range artifactUuidsToExpand {
		artifactFile, err := expander.filesArtifactStore.GetFileByUUID(string(filesArtifactId))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", filesArtifactId)
		}

		filesArtifactExpansionVolume, err := expander.kurtosisBackend.CreateFilesArtifactExpansionVolume(
			ctx,
			expander.enclaveId,
			serviceGUID,
			filesArtifactId,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion volume for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGUID, filesArtifactId, expander.enclaveId)
		}
		volumeName := filesArtifactExpansionVolume.GetName()
		volumesToDestroyIfSomethingFails[volumeName] = true
		defer func() {
			if len(volumesToDestroyIfSomethingFails) > 0 {
				expander.destroyFilesArtifactExpansionVolumes(ctx, volumesToDestroyIfSomethingFails)
				//We rewrite this var here to prevent more than one execution becase it is in a loop
				volumesToDestroyIfSomethingFails = map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool{}
			}
		}()

		if err := expander.runFilesArtifactExpander(
			ctx,
			filesArtifactId,
			serviceGUID,
			volumeName,
			artifactFile.GetFilepathRelativeToDataDirRoot(),
		); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running files artifact expander for user service with GUID '%v' and files artifact ID '%v' and files artifact expansion volume '%v' in enclave with ID '%v'",serviceGUID, filesArtifactId, volumeName, expander.enclaveId)
		}

		artifactIdsToVolNames[filesArtifactId] = volumeName
	}

	//We rewrite this var to avoid destroying them if everything is ok
	volumesToDestroyIfSomethingFails = map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool{}
	return artifactIdsToVolNames, nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
// NOTE: This is a separate function so we can defer the releasing of the IP address and guarantee that it always
//  goes back into the IP pool
func (expander *FilesArtifactExpander) runFilesArtifactExpander(
	ctx context.Context,
	filesArtifactId service.FilesArtifactID,
	serviceGuid service.ServiceGUID,
	filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName,
	artifactFilepathRelativeToEnclaveDataVolRoot string,
) error {
	// NOTE: This silently (temporarily) uses up one of the user's requested IP addresses with a node
	//  that's not one of their services! This could get confusing if the user requests exactly a wide enough
	//  subnet to fit all _their_ services, but we hit the limit because we have these admin containers too
	//  If this becomes problematic, create a special "admin" network, one per suite execution, for doing thinks like this?
	// TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	expanderIpAddr, err := expander.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a free IP for the files artifact expander")
	}
	defer expander.freeIpAddrTracker.ReleaseIpAddr(expanderIpAddr)

	guid := newFilesArtifactExpanderGUID(filesArtifactId, serviceGuid)

	if _, err := expander.kurtosisBackend.RunFilesArtifactExpander(
		ctx,
		guid,
		expander.enclaveId,
		filesArtifactExpansionVolumeName,
		destVolMntDirpathOnExpander,
		artifactFilepathRelativeToEnclaveDataVolRoot,
		expanderIpAddr,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred running files artifact expander with GUID '%v' for files artifact expansion volume '%v' in enclave with ID '%v'", guid, filesArtifactExpansionVolumeName, expander.enclaveId)
	}

	return nil
}

func (expander *FilesArtifactExpander) destroyFilesArtifactExpansionVolumes(ctx context.Context, volumeNamesSet map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool) {
	filesArtifactExpansionVolumeFilters := &files_artifact_expansion_volume.FilesArtifactExpansionVolumeFilters{
		Names: volumeNamesSet,
	}
	_, erroredVolumeNames, err := expander.kurtosisBackend.DestroyFilesArtifactExpansionVolumes(
		ctx,
		filesArtifactExpansionVolumeFilters)
	if err != nil || len(erroredVolumeNames) > 0 {
		var volumeNamesStr string
		var destroyVolumesErr error
		if err != nil {
			destroyVolumesErr = err
			for volumeName := range volumeNamesSet {
				volumeNameStr := string(volumeName)
				volumeNamesStr = strings.Join([]string{volumeNameStr}, ", ")
			}
		}
		if len(erroredVolumeNames) > 0 {
			volumeErrStrs := []string{}
			for volumeName, destroyVolErr := range erroredVolumeNames{
				volumeNameStr := string(volumeName)
				volumeNamesStr = strings.Join([]string{volumeNameStr}, ", ")
				volumeErrStr := fmt.Sprintf("An error occurred destroying files artifact expansion volume '%v':\n%v", volumeNameStr, destroyVolErr)
				volumeErrStrs = append(volumeErrStrs, volumeErrStr)
			}
			errorMsg := strings.Join(volumeErrStrs, "\n\n")
			destroyVolumesErr = stacktrace.NewError(errorMsg)
		}
		logrus.Error("Creating files artifact expansion volumes failed, but an error occurred destroying volumes we started:")
		fmt.Fprintln(logrus.StandardLogger().Out, destroyVolumesErr)
		logrus.Errorf("ACTION REQUIRED: You'll need to manually kill volumes with name '%v'", volumeNamesStr)
	}
}

func newFilesArtifactExpanderGUID(filesArtifactId service.FilesArtifactID, userServiceGuid service.ServiceGUID) files_artifact_expander.FilesArtifactExpanderGUID {
	userServiceGuidStr := string(userServiceGuid)
	filesArtifactIdStr := string(filesArtifactId)
	suffix := current_time_str_provider.GetCurrentTimeStr()
	guidStr := strings.Join([]string{userServiceGuidStr, filesArtifactIdStr, suffix}, guidElementSeparator)
	guid := files_artifact_expander.FilesArtifactExpanderGUID(guidStr)
	return guid
}
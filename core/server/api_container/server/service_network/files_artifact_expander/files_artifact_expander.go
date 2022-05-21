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

	filesArtifactStore *enclave_data_directory.FilesArtifactStore
}

func NewFilesArtifactExpander(kurtosisBackend backend_interface.KurtosisBackend, enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider, enclaveId enclave.EnclaveID, filesArtifactStore *enclave_data_directory.FilesArtifactStore) *FilesArtifactExpander {
	return &FilesArtifactExpander{kurtosisBackend: kurtosisBackend, enclaveObjAttrsProvider: enclaveObjAttrsProvider, enclaveId: enclaveId, filesArtifactStore: filesArtifactStore}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
	ctx context.Context,
	serviceGuid service.ServiceGUID, // GUID of the service for whom the artifacts are being expanded into volumes
	artifactUuidsToExpand map[service.FilesArtifactID]bool,
) (map[service.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, error) {

	// TODO PERF: parallelize this to increase speed
	artifactIdsToVolNames := map[service.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName{}
	for filesArtifactId := range artifactUuidsToExpand {
		artifactFile, err := expander.filesArtifactStore.GetFileByUUID(string(filesArtifactId))
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", filesArtifactId)
		}

		_, err = expander.kurtosisBackend.CreateFilesArtifactExpansion(
			ctx,
			expander.enclaveId,
			serviceGuid,
			filesArtifactId,
			artifactFile.GetFilepathRelativeToDataDirRoot())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, expander.enclaveId)
		}

		artifactIdsToVolNames[filesArtifactId] = ""
	}
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
	guid := newFilesArtifactExpanderGUID(filesArtifactId, serviceGuid)
	if _, err := expander.kurtosisBackend.RunFilesArtifactExpander(
		ctx,
		guid,
		expander.enclaveId,
		filesArtifactExpansionVolumeName,
		destVolMntDirpathOnExpander,
		artifactFilepathRelativeToEnclaveDataVolRoot,
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

func newFilesArtifactExpanderGUID(filesArtifactId service.FilesArtifactID, serviceGuid service.ServiceGUID) files_artifact_expander.FilesArtifactExpanderGUID {
	serviceRegistrationGuidStr := string(serviceGuid)
	filesArtifactIdStr := string(filesArtifactId)
	suffix := current_time_str_provider.GetCurrentTimeStr()
	guidStr := strings.Join([]string{serviceRegistrationGuidStr, filesArtifactIdStr, suffix}, guidElementSeparator)
	guid := files_artifact_expander.FilesArtifactExpanderGUID(guidStr)
	return guid
}
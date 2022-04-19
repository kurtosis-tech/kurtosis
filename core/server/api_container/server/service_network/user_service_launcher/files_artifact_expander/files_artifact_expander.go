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
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expander"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion_volume"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/current_time_str_provider"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"path"
	"strings"
)

const (
	// Dirpath on the artifact expander container where the enclave data dir (which contains the artifacts)
	//  will be bind-mounted
	enclaveDataDirMountpointOnExpanderContainer = "/enclave-data"

	// Dirpath on the artifact expander container where the destination volume containing expanded files will be mounted
	destVolMntDirpathOnExpander = "/dest"

	guidElementSeparator = "-"
)

/*
Class responsible for taking an artifact containing compressed files and creating a new
	files artifact expansion volume and a new the files artifact expander for it.
	The new files artifact expander will be on charge of uncompressing the artifact contents
	into the new volume, and this volume will be mounted on a new service
*/
type FilesArtifactExpander struct {
	// Host machine dirpath so the expander can bind-mount it to the artifact expansion containers
	enclaveDataDirpathOnHostMachine string

	kurtosisBackend backend_interface.KurtosisBackend

	enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider

	enclaveId enclave.EnclaveID

	freeIpAddrTracker *lib.FreeIpAddrTracker

	filesArtifactCache *enclave_data_directory.FilesArtifactCache
}

func NewFilesArtifactExpander(enclaveDataDirpathOnHostMachine string, kurtosisBackend backend_interface.KurtosisBackend, enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider, enclaveId enclave.EnclaveID, freeIpAddrTracker *lib.FreeIpAddrTracker, filesArtifactCache *enclave_data_directory.FilesArtifactCache) *FilesArtifactExpander {
	return &FilesArtifactExpander{enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine, kurtosisBackend: kurtosisBackend, enclaveObjAttrsProvider: enclaveObjAttrsProvider, enclaveId: enclaveId, freeIpAddrTracker: freeIpAddrTracker, filesArtifactCache: filesArtifactCache}
}

func (expander FilesArtifactExpander) ExpandArtifactsIntoVolumes(
	ctx context.Context,
	serviceGUID service.ServiceGUID, // Service GUID for whom the artifacts are being expanded into volumes
	artifactIdsToExpand map[files_artifact.FilesArtifactID]bool,
) (map[files_artifact.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName, error) {

	// TODO PERF: parallelize this to increase speed
	artifactIdsToVolNames := map[files_artifact.FilesArtifactID]files_artifact_expansion_volume.FilesArtifactExpansionVolumeName{}
	for filesArtifactId := range artifactIdsToExpand {
		artifactFile, err := expander.filesArtifactCache.GetFilesArtifact(filesArtifactId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", filesArtifactId)
		}

		artifactRelativeFilepath := artifactFile.GetFilepathRelativeToDataDirRoot()
		artifactFilepathOnExpanderContainer := path.Join(
			enclaveDataDirMountpointOnExpanderContainer,
			artifactRelativeFilepath,
		)

		filesArtifactExpansionVolume, err := expander.kurtosisBackend.CreateFilesArtifactExpansionVolume(
			ctx,
			expander.enclaveId,
			serviceGUID,
			filesArtifactId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion volume for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGUID, filesArtifactId, expander.enclaveId)
		}
		volumeName := filesArtifactExpansionVolume.GetName()

		if err := expander.runFilesArtifactExpander(
			ctx,
			filesArtifactId,
			serviceGUID,
			volumeName,
			artifactFilepathOnExpanderContainer,
		); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred running files artifact expander for user service with GUID '%v' and files artifact ID '%v' and files artifact expansion volume '%v' in enclave with ID '%v'",serviceGUID, filesArtifactId, volumeName, expander.enclaveId)
		}

		artifactIdsToVolNames[filesArtifactId] = volumeName
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
	filesArtifactId files_artifact.FilesArtifactID,
	serviceGuid service.ServiceGUID,
	filesArtifactExpansionVolumeName files_artifact_expansion_volume.FilesArtifactExpansionVolumeName,
	artifactFilepathOnExpanderContainer string,
) error {
	shouldDestroyVolume := true
	defer func() {
		if shouldDestroyVolume {
			filesArtifactExpansionVolumeFilters := &files_artifact_expansion_volume.FilesArtifactExpansionVolumeFilters{
				Names: map[files_artifact_expansion_volume.FilesArtifactExpansionVolumeName]bool{
					filesArtifactExpansionVolumeName:true,
				},
			}
			_, erroredVolumeNames, destroyVolumeErr := expander.kurtosisBackend.DestroyFilesArtifactExpansionVolumes(
				ctx,
				filesArtifactExpansionVolumeFilters)
			if destroyVolumeErr != nil || len(erroredVolumeNames) > 0 {
				if destroyVolumeErr == nil {
					destroyVolumeErr = erroredVolumeNames[filesArtifactExpansionVolumeName]
				}
				logrus.Error("Creating files artifact expansion volume failed, but an error occurred killing volume we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, destroyVolumeErr)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill volume with name '%v'", filesArtifactExpansionVolumeName)
			}
		}
	}()

	// NOTE: This silently (temporarily) uses up one of the user's requested IP addresses with a node
	//  that's not one of their services! This could get confusing if the user requests exactly a wide enough
	//  subnet to fit all _their_ services, but we hit the limit because we have these admin containers too
	//  If this becomes problematic, create a special "admin" network, one per suite execution, for doing thinks like this?
	// TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	expanderIrAddr, err := expander.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a free IP for the files artifact expander")
	}
	defer expander.freeIpAddrTracker.ReleaseIpAddr(expanderIrAddr)

	guid := newFilesArtifactExpanderGUID(filesArtifactId, serviceGuid)

	expander.kurtosisBackend.RunFilesArtifactExpander(
		ctx,
		guid,
		expander.enclaveId,
		filesArtifactExpansionVolumeName,
		expander.enclaveDataDirpathOnHostMachine,
		destVolMntDirpathOnExpander,
		artifactFilepathOnExpanderContainer,
		expanderIrAddr,
	)

	shouldDestroyVolume = false
	return nil
}

func newFilesArtifactExpanderGUID(filesArtifactId files_artifact.FilesArtifactID, userServiceGuid service.ServiceGUID) files_artifact_expander.FilesArtifactExpanderGUID {
	userServiceGuidStr := string(userServiceGuid)
	filesArtifactIdStr := string(filesArtifactId)
	suffix := current_time_str_provider.GetCurrentTimeStr()
	guidStr := strings.Join([]string{userServiceGuidStr, filesArtifactIdStr, suffix}, guidElementSeparator)
	guid := files_artifact_expander.FilesArtifactExpanderGUID(guidStr)
	return guid
}

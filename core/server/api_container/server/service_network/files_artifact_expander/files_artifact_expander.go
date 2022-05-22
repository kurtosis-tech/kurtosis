/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package files_artifact_expander

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
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

func (expander FilesArtifactExpander) ExpandArtifacts(
	ctx context.Context,
	serviceGuid service.ServiceGUID, // GUID of the service for whom the artifacts are being expanded into volumes
	artifactUuidsToExpand map[service.FilesArtifactID]bool,
) (map[service.FilesArtifactID]files_artifact_expansion.FilesArtifactExpansionGUID, error) {


	// TODO PERF: parallelize this to increase speed
	artifactIdsToExpansionGUIDs := map[service.FilesArtifactID]files_artifact_expansion.FilesArtifactExpansionGUID{}
	for filesArtifactId := range artifactUuidsToExpand {
		artifactFile, err := expander.filesArtifactStore.GetFile(filesArtifactId)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the file for files artifact '%v'", filesArtifactId)
		}

		expansionGUID, err := expander.kurtosisBackend.CreateFilesArtifactExpansion(
			ctx,
			expander.enclaveId,
			serviceGuid,
			artifactFile.GetFilepathRelativeToDataDirRoot(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating files artifact expansion for user service with GUID '%v' and files artifact ID '%v' in enclave with ID '%v'", serviceGuid, filesArtifactId, expander.enclaveId)
		}
		artifactIdsToExpansionGUIDs[filesArtifactId] = expansionGUID.GetGUID()
	}
	return artifactIdsToExpansionGUIDs, nil
}

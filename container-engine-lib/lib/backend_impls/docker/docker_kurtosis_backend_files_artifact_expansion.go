package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *DockerKurtosisBackend) CreateFilesArtifactExpansion(ctx context.Context, enclaveId enclave.EnclaveID, serviceGuid service.ServiceGUID, filesArtifactId service.FilesArtifactID, filesArtifactFilepathRelativeToEnclaveDatadirRoot string, ) (*files_artifact_expansion.FilesArtifactExpansionGUID, error, ) {
	artifactExpansionVolume, err := backend.CreateFilesArtifactExpansionVolume(ctx, enclaveId, serviceGuid, filesArtifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create files artifact expansion volume.")
	}
	logrus.Infof("'%v'", artifactExpansionVolume.GetName())
	return nil, nil
}

//Destroy files artifact expansion volume and expander using the given filters
func (backend *DockerKurtosisBackend)  DestroyFilesArtifactExpansion(
	ctx context.Context,
	filters  files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	panic("IMPLEMENT ME")
}

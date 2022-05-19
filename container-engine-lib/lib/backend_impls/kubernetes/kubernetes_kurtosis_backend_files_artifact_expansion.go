package kubernetes

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/files_artifact_expansion"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
)

//Create a files artifact exansion volume for user service and file artifact id and runs a file artifact expander
func (backend *KubernetesKurtosisBackend) CreateFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	serviceGuid service.ServiceGUID,
	filesArtifactId service.FilesArtifactID,
	filesArtifactFilepathRelativeToEnclaveDatadirRoot string)(*files_artifact_expansion.FilesArtifactExpansion, error) {
	panic("IMPLEMENT ME")
}

//Destroy files artifact expansion volume and expander using the given filters
func (backend *KubernetesKurtosisBackend)  DestroyFilesArtifactExpansion(
	ctx context.Context,
	enclaveId enclave.EnclaveID,
	filters  files_artifact_expansion.FilesArtifactExpansionFilters,
)(
	successfulFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]bool,
	erroredFileArtifactExpansionGUIDs map[files_artifact_expansion.FilesArtifactExpansionGUID]error,
	resultErr error,
) {
	panic("IMPLEMENT ME")
}

package services_files_artifacts

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	uuidKey                                       = "uuid"
	serviceDirPathsToFilesArtifactsIdentifiersKey = "services-to-files-artifacts"
)

type servicesFilesArtifacts struct {
	uuid service.ServiceUUID

	serviceDirPathsToFilesArtifactIdentifiers map[string]string
}

func NewServicesFilesArtifactsObj(uuid service.ServiceUUID, serviceDirPathsToFilesArtifactsIdentifiers map[string]string) *servicesFilesArtifacts {
	return &servicesFilesArtifacts{uuid: uuid, serviceDirPathsToFilesArtifactIdentifiers: serviceDirPathsToFilesArtifactsIdentifiers}
}

func (servicesFilesArtifacts *servicesFilesArtifacts) GetUuid() service.ServiceUUID {
	return servicesFilesArtifacts.uuid
}

func (servicesFilesArtifacts *servicesFilesArtifacts) GetServiceDirPathsToFilesArtifactsIdentifiers() map[string]string {
	return servicesFilesArtifacts.serviceDirPathsToFilesArtifactIdentifiers
}

func (servicesFilesArtifacts *servicesFilesArtifacts) MarshalJSON() ([]byte, error) {

	data := map[string]string{}

	return json.Marshal(data)
}

func (servicesFilesArtifacts *servicesFilesArtifacts) UnmarshalJSON(data []byte) error {

	unmarshalledMapPtr := &map[string]string{}

	if err := json.Unmarshal(data, unmarshalledMapPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling map")
	}

	return nil
}

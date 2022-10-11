package docker_labels_for_logs

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
)

//The following docker labels will be added into the logs stream which is necessary for creating new tags
//in the logs database and then use it for querying them to get the specific user service's logs
var LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream = []*docker_label_key.DockerLabelKey{
	label_key_consts.GUIDDockerLabelKey,
	label_key_consts.ContainerTypeDockerLabelKey,
}

//The 'enclaveID' value is used for Fluentbit to send it to Loki as the "X-Scope-OrgID" request's header
//due Loki is now configured to use multi tenancy, and we established this relation: enclaveID = tenantID
var LogsDatabaseKurtosisTrackedDockerLabelUsedForIdentifyTenants = label_key_consts.EnclaveIDDockerLabelKey

//These are all the logs database Kurtosis tracked Docker Labels used
func GetAllLogsDatabaseKurtosisTrackedDockerLabels() []*docker_label_key.DockerLabelKey {
	allLogsDatabaseKurtosisTrackedDockerLabelsSet := LogsDatabaseKurtosisTrackedDockerLabelsForIdentifyLogsStream
	allLogsDatabaseKurtosisTrackedDockerLabelsSet = append(allLogsDatabaseKurtosisTrackedDockerLabelsSet, LogsDatabaseKurtosisTrackedDockerLabelUsedForIdentifyTenants)
	return allLogsDatabaseKurtosisTrackedDockerLabelsSet
}

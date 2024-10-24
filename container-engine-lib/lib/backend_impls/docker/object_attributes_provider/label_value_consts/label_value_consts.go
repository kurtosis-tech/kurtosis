package label_value_consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	appIdLabelValueStr                       = "kurtosis"
	engineContainerTypeLabelValueStr         = "kurtosis-engine"
	logsAggregatorContainerTypeLabelValueStr = "kurtosis-logs-aggregator"
	logsCollectorContainerTypeLabelValueStr  = "kurtosis-logs-collector"
	reverseProxyContainerTypeLabelValueStr   = "kurtosis-reverse-proxy"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	apiContainerContainerTypeLabelValueStr           = "api-container"
	userServiceContainerTypeLabelValueStr            = "user-service"
	filesArtifactsExpanderContainerTypeLabelValueStr = "files-artifacts-expander"

	enclaveDataVolumeTypeLabelValueStr            = "enclave-data"
	filesArtifactExpansionVolumeTypeLabelValueStr = "files-artifacts-expansion"
	persistentDirectoryVolumeTypeLabelValueStr    = "persistent-directory"
	logsStorageVolumeTypeLabelValueStr            = "kurtosis-logs-storage"
	logsCollectorVolumeTypeLabelValueStr          = "logs-collector-data"
	githubAuthStorageVolumeTypeLabelValueStr      = "github-auth-storage"
	dockerConfigStorageVolumeTypeLabelValueStr    = "docker-config-storage"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//
//	which will cause a resource leak on the user's system!
//
//	 If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var AppIDDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(appIdLabelValueStr)
var EngineContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)
var LogsAggregatorTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(logsAggregatorContainerTypeLabelValueStr)
var LogsCollectorTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(logsCollectorContainerTypeLabelValueStr)
var ReverseProxyTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(reverseProxyContainerTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var APIContainerContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(apiContainerContainerTypeLabelValueStr)
var UserServiceContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(userServiceContainerTypeLabelValueStr)
var FilesArtifactExpanderContainerTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactsExpanderContainerTypeLabelValueStr)

var EnclaveDataVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(enclaveDataVolumeTypeLabelValueStr)
var FilesArtifactExpansionVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(filesArtifactExpansionVolumeTypeLabelValueStr)
var PersistentDirectoryVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(persistentDirectoryVolumeTypeLabelValueStr)
var LogsStorageVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(logsStorageVolumeTypeLabelValueStr)
var LogsCollectorVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(logsCollectorVolumeTypeLabelValueStr)
var GitHubAuthStorageVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(githubAuthStorageVolumeTypeLabelValueStr)
var DockerConfigStorageVolumeTypeDockerLabelValue = docker_label_value.MustCreateNewDockerLabelValue(dockerConfigStorageVolumeTypeLabelValueStr)

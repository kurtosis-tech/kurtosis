package vector

import (
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
)

const (
	vectorContainerName = "vector"
	vectorImage         = "timberio/vector:0.45.0-debian"

	binaryFilepath = "/usr/bin/vector"

	vectorConfigVolumeName = "vector-config"
	vectorConfigMountPath  = "/etc/vector"
	vectorConfigFileName   = "vector.yaml"
	vectorConfigFilePath   = vectorConfigMountPath + "/" + vectorConfigFileName

	kurtosisLogsVolumeName = "varlogskurtosis"
	kurtosisLogsMountPath  = "/var/log/kurtosis"

	apiPort = 8686

	// mount the data directory as the disk buffer for file sink is contained here and needs to be persisted onto the k8s node in case vector restarts
	vectorDataDirVolumeName = "varlibvector"
	vectorDataDirMountPath  = "/var/lib/vector"

	bufferSizeStr = "268435488" // 256 MB is min for vector

	defaultSourceId          = "kurtosis_default_source"
	fluentBitSourceType      = "fluent"
	fluentBitSourceIpAddress = "0.0.0.0"
	fileSinkType             = "file"
)

var (
	uuidLogsFilepath = fmt.Sprintf("\"%s/%%G/%%V/{{ %v }}/{{ %v }}.json\"", kurtosisLogsMountPath, kubernetes_label_key.LogsEnclaveUUIDKubernetesLabelKey.GetString(), kubernetes_label_key.LogsServiceUUIDKubernetesLabelKey.GetString())
)

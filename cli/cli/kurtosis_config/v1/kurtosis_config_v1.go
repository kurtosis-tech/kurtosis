package v1

import "github.com/kurtosis-tech/stacktrace"

const (
	kurtosisConfigV1DockerType     = "docker"
	kurtosisConfigV1KubernetesType = "kubernetes"
)

// NOTE: All new YAML property names here should be kebab-case because
//a) it's easier to read b) it's easier to write
//c) it's consistent with previous properties and changing the format of
//an already-written config file is very difficult

const (
	versionNumber = 1
	defaultDockerClusterName = "local-docker"
	defaultMinikubeClusterName = "minikube"
	defaultMinikubeStorageClass = "standard"
	defaultMinikubeGigabytesPerEnclave = 10
)

type KurtosisConfigV1 struct {
	//We set public fields because YAML marshalling needs it on this way
	//All fields should be pointers, that way we can enforce required fields
	//by detecting nil pointers.
	ConfigVersion *int `yaml:"config-version"`
	ShouldSendMetrics *bool                         `yaml:"should-send-metrics"`
	KurtosisClusters *map[string]*KurtosisClusterV1 `yaml:"kurtosis-clusters"`
}

func NewDefaultKurtosisConfigV1() *KurtosisConfigV1 {
	version := versionNumber
	dockerClusterConfig := getDefaultDockerKurtosisClusterConfig()
	minikubeClusterConfig := getDefaultMinikubeKurtosisClusterConfig()
	kurtosisClusters := map[string]*KurtosisClusterV1{
		defaultDockerClusterName:   dockerClusterConfig,
		defaultMinikubeClusterName: minikubeClusterConfig,
	}
	return &KurtosisConfigV1{
		ConfigVersion: &version,
		KurtosisClusters: &kurtosisClusters,
	}
}

func (kurtosisConfigV1 *KurtosisConfigV1) Validate() error {
	if kurtosisConfigV1.ShouldSendMetrics == nil {
		return stacktrace.NewError("ShouldSendMetrics field of Kurtosis Config v1 is nil, when it should be true or false.")
	}
	if kurtosisConfigV1.ConfigVersion == nil {
		return stacktrace.NewError("ConfigVersion field of Kurtosis Config v1 is nil, when it should be %d.", versionNumber)
	}
	if *kurtosisConfigV1.ConfigVersion != versionNumber {
		return stacktrace.NewError("ConfigVersion field of Kurtosis Config v1 is %d, when it should be %d.", kurtosisConfigV1.ConfigVersion, versionNumber)
	}
	if kurtosisConfigV1.KurtosisClusters == nil {
		return stacktrace.NewError("KurtosisCluster field of Kurtosis Config v1 is nil, when it should have a map of Kurtosis cluster configurations.")
	}
	if len(*kurtosisConfigV1.KurtosisClusters) == 0 {
		return stacktrace.NewError("KurtosisCluster field of Kurtosis Config v1 has no clusters, when it should have at least one.")
	}
	for clusterId, clusterConfig := range *kurtosisConfigV1.KurtosisClusters {
		clusterValid := clusterConfig.Validate(clusterId)
		if clusterValid != nil {
			return clusterValid
		}
	}
	return nil
}

func (kurtosisConfigV1 *KurtosisConfigV1) OverlayOverrides(overrides *KurtosisConfigV1) {
	if overrides.ShouldSendMetrics != nil {
		kurtosisConfigV1.ShouldSendMetrics = overrides.ShouldSendMetrics
	}
	if overrides.KurtosisClusters != nil {
		for clusterId, clusterConfig := range *overrides.KurtosisClusters {
			(*kurtosisConfigV1.KurtosisClusters)[clusterId] = clusterConfig
		}
	}
}

type KurtosisClusterV1 struct {
	Type *string                      `yaml:"type"`
	Config *KubernetesClusterConfigV1 `yaml:"config"`
}

func (kurtosisClusterV1 *KurtosisClusterV1) Validate(clusterId string) error {
	clusterConfig := kurtosisClusterV1
	if clusterConfig.Type == nil {
		return stacktrace.NewError("KurtosisCluster '%v' has nil Type field, when it should be %v or %v",
			clusterId, kurtosisConfigV1DockerType, kurtosisConfigV1KubernetesType)
	}
	if *clusterConfig.Type != kurtosisConfigV1DockerType && *clusterConfig.Type != kurtosisConfigV1KubernetesType {
		return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v', when it should be %v or %v",
			clusterId, *clusterConfig.Type, kurtosisConfigV1DockerType, kurtosisConfigV1KubernetesType)
	}
	if *clusterConfig.Type == kurtosisConfigV1KubernetesType {
		if clusterConfig.Config == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no Config field. Config fields are required for type %v",
				clusterId, *clusterConfig.Type, kurtosisConfigV1KubernetesType)
		}
		if clusterConfig.Config.KubernetesClusterName == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no Kubernetes cluster name in its config map.",
				clusterId, *clusterConfig.Type)
		}
		if clusterConfig.Config.StorageClass == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no StorageClass name in its config map.",
				clusterId, *clusterConfig.Type)
		}
		if clusterConfig.Config.EnclaveSizeInGigabytes == nil {
			return stacktrace.NewError("KurtosisCluster '%v' has Type field '%v' but has no EnclaveSizeInGigabytes specified in its config map.",
				clusterId, *clusterConfig.Type)
		}
	}
	return nil
}

type KubernetesClusterConfigV1 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name"`
	StorageClass *string `yaml:"storage-class"`
	EnclaveSizeInGigabytes *int `yaml:"enclave-size-in-gigabytes"`
}

// ===================== HELPERS==============================

func getDefaultDockerKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := kurtosisConfigV1DockerType
	return &KurtosisClusterV1{
		Type: &clusterType,
	}
}

func getDefaultMinikubeKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := kurtosisConfigV1KubernetesType
	minikubeKubernetesCluster := getDefaultMinikubeKubernetesClusterConfig()
	return &KurtosisClusterV1{
		Type: &clusterType,
		Config: minikubeKubernetesCluster,
	}
}

func getDefaultMinikubeKubernetesClusterConfig() *KubernetesClusterConfigV1 {
	kubernetesClusterName := defaultMinikubeClusterName
	storageClass := defaultMinikubeStorageClass
	gbPerEnclave := defaultMinikubeGigabytesPerEnclave
	clusterConfig := KubernetesClusterConfigV1{
		KubernetesClusterName: &kubernetesClusterName,
		StorageClass: &storageClass,
		EnclaveSizeInGigabytes: &gbPerEnclave,
	}
	return &clusterConfig
}



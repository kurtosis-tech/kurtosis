package kurtosis_config
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
	ShouldSendMetrics *bool `yaml:"should-send-metrics"`
	KurtosisClusters *map[string]*KurtosisClusterV1 `yaml:"kurtosis-clusters"`
}

type KurtosisClusterV1 struct {
	Type *string `yaml:"type"`
	Config *KubernetesClusterConfigV1 `yaml:"config"`
}

type KubernetesClusterConfigV1 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name"`
	StorageClass *string `yaml:"storage-class"`
	EnclaveSizeInGigabytes *int `yaml:"enclave-size-in-gigabytes"`
}

func getDefaultDockerKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := KurtosisConfigV1DockerType
	return &KurtosisClusterV1{
		Type: &clusterType,
	}
}

func getDefaultMinikubeKurtosisClusterConfig() *KurtosisClusterV1 {
	clusterType := KurtosisConfigV1KubernetesType
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

func NewDefaultKurtosisConfigV1(doesUserAcceptSendingMetrics *bool) *KurtosisConfigV1 {
	version := versionNumber
	dockerClusterConfig := getDefaultDockerKurtosisClusterConfig()
	minikubeClusterConfig := getDefaultMinikubeKurtosisClusterConfig()
	kurtosisClusters := map[string]*KurtosisClusterV1{
		defaultDockerClusterName: dockerClusterConfig,
		defaultMinikubeClusterName: minikubeClusterConfig,
	}
	return &KurtosisConfigV1{
		ConfigVersion: &version,
		ShouldSendMetrics: doesUserAcceptSendingMetrics,
		KurtosisClusters: &kurtosisClusters,
	}
}


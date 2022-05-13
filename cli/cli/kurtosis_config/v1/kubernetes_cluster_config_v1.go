package v1

type KubernetesClusterConfigV1 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name,omitempty"`
	StorageClass *string `yaml:"storage-class,omitempty"`
	EnclaveSizeInGigabytes *uint `yaml:"enclave-size-in-gigabytes,omitempty"`
}


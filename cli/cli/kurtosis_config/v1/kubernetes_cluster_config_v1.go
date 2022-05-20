package v1

type KubernetesClusterConfigV1 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name,omitempty"`
	StorageClass *string `yaml:"storage-class,omitempty"`
	EnclaveSizeInMegabytes *uint `yaml:"enclave-size-in-Megabytes,omitempty"`
}


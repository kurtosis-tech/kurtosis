package v2

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

type KubernetesClusterConfigV2 struct {
	KubernetesClusterName *string `yaml:"kubernetes-cluster-name,omitempty"`
	StorageClass *string `yaml:"storage-class,omitempty"`
	EnclaveSizeInMegabytes *uint `yaml:"enclave-size-in-megabytes,omitempty"`
}


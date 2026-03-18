package v7

/*
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
                           DO NOT CHANGE THIS FILE!
  If you change this file, it will break config for users who have instantiated an
           overrides file with this version of config overrides!
    Instead, to make changes, you will need to add a new version of the config
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
*/

// KubernetesTolerationV7 represents a Kubernetes toleration for pod scheduling.
type KubernetesTolerationV7 struct {
	Key               *string `yaml:"key,omitempty"`
	Operator          *string `yaml:"operator,omitempty"`
	Value             *string `yaml:"value,omitempty"`
	Effect            *string `yaml:"effect,omitempty"`
	TolerationSeconds *int64  `yaml:"toleration-seconds,omitempty"`
}

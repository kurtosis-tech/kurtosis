package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/versioned_config"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	defaultDockerClusterName = "docker"

	defaultMinikubeClusterName = "minikube"

	defaultMinikubeClusterKubernetesClusterNameStr = "minikube"
	defaultMinikubeStorageClass = "standard"
	defaultMinikubeEnclaveDataVolumeGb = uint(2)
)

/*
	KurtosisConfig should be the interface other modules use to access
	the latest configuration values available in Kurtosis CLI configuration.

	From the standpoint of the rest of our code, this is the evergreen config value.
	This prevents code using configuration from needing to completely change
	everytime configuration versions change.

	Under the hood, the KurtosisConfig is responsible for reconciling the user's overrides
	with the default values for the configuration. It can be thought of as a "resolver" for
	the overrides on top of the default config.
 */
type KurtosisConfig struct {
	// Only necessary to store for when we serialize overrides
	overrides      *v1.KurtosisConfigV1

	shouldSendMetrics bool
	clusters map[string]*KurtosisClusterConfig
}

// Constructs a new KurtosisConfig that uses the given overrides
func NewKurtosisConfigFromOverrides(overrides *v1.KurtosisConfigV1) (*KurtosisConfig, error) {
	config := &KurtosisConfig{
		overrides: overrides,
	}

	// Get latest config version
	latestConfigVersion := config_version.ConfigVersion_v0
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	// Ensure that the overrides are storing the latest config version
	// From this point onwards, it should be impossible to not have the right config version
	config.overrides.ConfigVersion = &latestConfigVersion

	// --------------------- Validation --------------------------
	if overrides.ShouldSendMetrics == nil {
		return nil, stacktrace.NewError("An explicit election about sending metrics must be made")
	}
	shouldSendMetrics := *overrides.ShouldSendMetrics



	for clusterId, clusterConfig := range overrides.KurtosisClusters {

	}

	return &KurtosisConfig{
		overrides:         nil,
		shouldSendMetrics: config.overrides.,
		clusters:          nil,
	}, nil
}

// NOTE: We probably want to remove this function entirely
func NewKurtosisConfigFromRequiredFields(didUserAcceptSendingMetrics bool) (*KurtosisConfig, error) {
	overrides := &v1.KurtosisConfigV1{
		ShouldSendMetrics: &didUserAcceptSendingMetrics,
	}
	result, err := NewKurtosisConfigFromOverrides(overrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kurtosis config with did-accept-metrics flag '%v'", didUserAcceptSendingMetrics)
	}
	return result, nil
}

/*
func (kurtosisConfig *KurtosisConfig) WithOverrides(overrides *v1.KurtosisConfigV1) (*KurtosisConfig, error) {
	renderedKurtosisConfig, err := kurtosisConfig.renderedConfig.OverlayOverrides(overrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to overlay configuration overrides on default Kurtosis configuration.")
	}
	return &KurtosisConfig{
		renderedConfig: renderedKurtosisConfig,
		overrides:      overrides,
	}, nil
}
 */

func (kurtosisConfig *KurtosisConfig) GetOverrides() *v1.KurtosisConfigV1 {
	return kurtosisConfig.overrides
}

func (kurtosisConfig *KurtosisConfig) GetConfigVersion() config_version.ConfigVersion {
	return kurtosisConfig
	if kurtosisConfig.overrides.ConfigVersion == nil {
		// This is a required field, and our validation should force that this is non-nil
		panic("The config overrides did not specify a config version; this should never happen")
	}
	return *kurtosisConfig.overrides.ConfigVersion
}

func (kurtosisConfig *KurtosisConfig) GetShouldSendMetrics() bool {
	return kurtosisConfig.shouldSendMetrics
}

func (kurtosisConfig *KurtosisConfig) GetKurtosisClusters() map[string]*KurtosisClusterConfig {
	return kurtosisConfig.clusters
}

func (kurtosisConfig *KurtosisConfig) Validate() error {
	overrides := kurtosisConfig.overrides

	// ------------------------------- Check required fields --------------------------------------
	if overrides.ConfigVersion == nil {
		// Should never happen
		return stacktrace.NewError("The Kurtosis config doesn't have a version; this is a bug in Kurtosis and should " +
			"never happen because we set the config version at construction time")
	}

	// ------------------------------- Check rest of state --------------------------------------


	return nil
}

func getDefaultKurtosisClusterConfigOverrides() map[string]*v1.KurtosisClusterV1 {
	dockerClusterType := KurtosisClusterType_Docker.String()
	minikubeClusterType := KurtosisClusterType_Kubernetes.String()
	minikubeKubernetesClusterName := defaultMinikubeClusterKubernetesClusterNameStr
	minikubeStorageClass := defaultMinikubeStorageClass
	minikubeEnclaveDataVolSizeGb := defaultMinikubeEnclaveDataVolumeGb

	result := map[string]*v1.KurtosisClusterV1{
		defaultDockerClusterName: {
			Type:   &dockerClusterType,
			Config: nil, // Must be nil for Docker
		},
		defaultMinikubeClusterName: {
			Type:   &minikubeClusterType,
			Config: &v1.KubernetesClusterConfigV1{
				KubernetesClusterName:  &minikubeKubernetesClusterName,
				StorageClass:           &minikubeStorageClass,
				EnclaveSizeInGigabytes: &minikubeEnclaveDataVolSizeGb,
			},
		},
	}

	return result
}

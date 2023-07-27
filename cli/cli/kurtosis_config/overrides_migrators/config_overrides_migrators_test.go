package overrides_migrators

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOverridesMigratorsCompletenessTest(t *testing.T) {
	// Dynamically get latest config version
	var latestConfigVersion config_version.ConfigVersion
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	for _, configVersion := range config_version.ConfigVersionValues() {
		if configVersion == latestConfigVersion {
			continue
		}
		_, found := AllConfigOverridesMigrators[configVersion]
		require.True(t, found, "No config overrides migrator found for config version '%v'; you'll need to add one", configVersion.String())
	}
	numMigrators := len(AllConfigOverridesMigrators)
	numConfigVersions := len(config_version.ConfigVersionValues())
	require.Equal(
		t,
		numConfigVersions-1,
		numMigrators,
		"There are %v Kurtosis config versions but %v config overrides migrators were declared; this likely means "+
			"extra migrators were declared that shouldn't be (there should always be migrators = num_versions - 1)",
	)
}

//func TestOverridesMigratorsMigrateFromV2toV3(t *testing.T) {
//	newTrue := true
//	newDocker := "docker"
//	newKubernetes := "kubernetes"
//	newKubernetesClusterName := "minikube"
//	newStorageClass := "standard"
//	newEnclaveSizeInMegabytes := uint(10)
//
//	kurtosisClusters := map[string]*v2.KurtosisClusterConfigV2{}
//	kurtosisClusters["docker"] = &v2.KurtosisClusterConfigV2{
//		Type:   &newDocker,
//		Config: nil,
//	}
//
//	kurtosisClusters["minikube"] = &v2.KurtosisClusterConfigV2{
//		Type: &newKubernetes,
//		Config: &v2.KubernetesClusterConfigV2{
//			KubernetesClusterName:  &newKubernetesClusterName,
//			StorageClass:           &newStorageClass,
//			EnclaveSizeInMegabytes: &newEnclaveSizeInMegabytes,
//		},
//	}
//
//	k := &v2.KurtosisConfigV2{
//		ConfigVersion:     1,
//		ShouldSendMetrics: &newTrue,
//		KurtosisClusters:  kurtosisClusters,
//	}
//
//	result, err := migrateFromV2(k)
//	casted := result.(*v3.KurtosisConfigV3)
//	require.NoError(t, err)
//
//	require.Nil(t, casted.KurtosisClusters["docker"].Config)
//	require.NotNil(t, casted.KurtosisClusters["minikube"].Config)
//}

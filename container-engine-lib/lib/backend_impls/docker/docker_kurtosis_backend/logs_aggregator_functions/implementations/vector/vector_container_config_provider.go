package vector

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
)

type vectorContainerConfigProvider struct {
	config *VectorConfig
}

func newVectorContainerConfigProvider(config *VectorConfig) *vectorContainerConfigProvider {
	return &vectorContainerConfigProvider{config: config}
}

func (vector *vectorContainerConfigProvider) GetContainerArgs(
	containerName string,
	containerLabels map[string]string,
	logsAggregatorVolumeName string,
	networkId string,
) (*docker_manager.CreateAndStartContainerArgs, error) {

	// ports? http? tcp?

	// config?

	volumeMounts := map[string]string{
		logsAggregatorVolumeName: configDirpath,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		containerImage,
		containerName,
		networkId,
	).WithLabels(
		containerLabels,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	return createAndStartArgs, nil
}

func (vector *vectorContainerConfigProvider) getConfigContent() (string, error) {
	vectorConfigYAMLContent, err := yaml.Marshal(vector.config)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred marshalling Vector config '%+v'", vector.config)
	}
	vectorConfigYAMLContentStr := string(vectorConfigYAMLContent)
	return vectorConfigYAMLContentStr, nil
}

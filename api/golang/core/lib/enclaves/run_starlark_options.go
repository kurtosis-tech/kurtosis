package enclaves

import "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"

const (
	defaultRelativePathToMainFile = ""
	defaultMainFunctionName       = ""
	defaultSerializedParams       = "{}"
	defaultDryRun                 = false
	defaultParallelism            = 4
	defaultCloudInstanceId        = ""
	defaultCloudUserId            = ""
)

var defaultExperimentalFeatureFlags = []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag(nil)

type RunStarlarkConfig struct {
	relativePathToMainFile   string
	mainFunctionName         string
	serializedParams         string
	dryRun                   bool
	parallelism              int32
	experimentalFeatureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag
	cloudInstanceId          string
	cloudUserId              string
}

type RunStarlarkConfigOption func(*RunStarlarkConfig)

func NewRunStarlarkConfig(opts ...RunStarlarkConfigOption) *RunStarlarkConfig {
	config := &RunStarlarkConfig{
		relativePathToMainFile:   defaultRelativePathToMainFile,
		mainFunctionName:         defaultMainFunctionName,
		serializedParams:         defaultSerializedParams,
		dryRun:                   defaultDryRun,
		parallelism:              defaultParallelism,
		experimentalFeatureFlags: defaultExperimentalFeatureFlags,
		cloudInstanceId:          defaultCloudUserId,
		cloudUserId:              defaultCloudInstanceId,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func WithRelativePathToMainFile(relativePathToMainFile string) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.relativePathToMainFile = relativePathToMainFile
	}
}

func WithMainFunctionName(mainFunctionName string) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.mainFunctionName = mainFunctionName
	}
}

func WithSerializedParams(serializedParams string) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.serializedParams = serializedParams
	}
}

func WithDryRun(dryRun bool) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.dryRun = dryRun
	}
}

func WithParallelism(parallelism int32) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.parallelism = parallelism
	}
}

func WithExperimentalFeatureFlags(experimentalFeatureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.experimentalFeatureFlags = experimentalFeatureFlags
	}
}

func WithCloudInstanceId(cloudInstanceId string) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.cloudInstanceId = cloudInstanceId
	}
}

func WithCloudUserId(cloudUserId string) RunStarlarkConfigOption {
	return func(config *RunStarlarkConfig) {
		config.cloudUserId = cloudUserId
	}
}

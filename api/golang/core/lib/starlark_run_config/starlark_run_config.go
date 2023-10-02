package starlark_run_config

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

type StarlarkRunConfig struct {
	RelativePathToMainFile   string
	MainFunctionName         string
	SerializedParams         string
	DryRun                   bool
	Parallelism              int32
	ExperimentalFeatureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag
	CloudInstanceId          string
	CloudUserId              string
}

type StarlarkRunConfigOption func(*StarlarkRunConfig)

func NewRunStarlarkConfig(opts ...StarlarkRunConfigOption) *StarlarkRunConfig {
	config := &StarlarkRunConfig{
		RelativePathToMainFile:   defaultRelativePathToMainFile,
		MainFunctionName:         defaultMainFunctionName,
		SerializedParams:         defaultSerializedParams,
		DryRun:                   defaultDryRun,
		Parallelism:              defaultParallelism,
		ExperimentalFeatureFlags: defaultExperimentalFeatureFlags,
		CloudInstanceId:          defaultCloudUserId,
		CloudUserId:              defaultCloudInstanceId,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func WithRelativePathToMainFile(relativePathToMainFile string) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.RelativePathToMainFile = relativePathToMainFile
	}
}

func WithMainFunctionName(mainFunctionName string) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.MainFunctionName = mainFunctionName
	}
}

func WithSerializedParams(serializedParams string) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.SerializedParams = serializedParams
	}
}

func WithDryRun(dryRun bool) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.DryRun = dryRun
	}
}

func WithParallelism(parallelism int32) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.Parallelism = parallelism
	}
}

func WithExperimentalFeatureFlags(experimentalFeatureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.ExperimentalFeatureFlags = experimentalFeatureFlags
	}
}

func WithCloudInstanceId(cloudInstanceId string) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.CloudInstanceId = cloudInstanceId
	}
}

func WithCloudUserId(cloudUserId string) StarlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.CloudUserId = cloudUserId
	}
}

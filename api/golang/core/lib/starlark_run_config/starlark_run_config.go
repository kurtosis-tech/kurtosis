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
	defaultImageDownload          = kurtosis_core_rpc_api_bindings.ImageDownloadMode_missing
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
	ImageDownload            kurtosis_core_rpc_api_bindings.ImageDownloadMode
}

type starlarkRunConfigOption func(*StarlarkRunConfig)

func NewRunStarlarkConfig(opts ...starlarkRunConfigOption) *StarlarkRunConfig {
	config := &StarlarkRunConfig{
		RelativePathToMainFile:   defaultRelativePathToMainFile,
		MainFunctionName:         defaultMainFunctionName,
		SerializedParams:         defaultSerializedParams,
		DryRun:                   defaultDryRun,
		Parallelism:              defaultParallelism,
		ExperimentalFeatureFlags: defaultExperimentalFeatureFlags,
		CloudInstanceId:          defaultCloudInstanceId,
		CloudUserId:              defaultCloudUserId,
		ImageDownload:            defaultImageDownload,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

func WithRelativePathToMainFile(relativePathToMainFile string) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.RelativePathToMainFile = relativePathToMainFile
	}
}

func WithMainFunctionName(mainFunctionName string) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.MainFunctionName = mainFunctionName
	}
}

func WithSerializedParams(serializedParams string) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.SerializedParams = serializedParams
	}
}

func WithDryRun(dryRun bool) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.DryRun = dryRun
	}
}

func WithParallelism(parallelism int32) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.Parallelism = parallelism
	}
}

func WithExperimentalFeatureFlags(experimentalFeatureFlags []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.ExperimentalFeatureFlags = experimentalFeatureFlags
	}
}

func WithCloudInstanceId(cloudInstanceId string) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.CloudInstanceId = cloudInstanceId
	}
}

func WithCloudUserId(cloudUserId string) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.CloudUserId = cloudUserId
	}
}

func WithImageDownloadMode(imageDownloadMode kurtosis_core_rpc_api_bindings.ImageDownloadMode) starlarkRunConfigOption {
	return func(config *StarlarkRunConfig) {
		config.ImageDownload = imageDownloadMode
	}
}

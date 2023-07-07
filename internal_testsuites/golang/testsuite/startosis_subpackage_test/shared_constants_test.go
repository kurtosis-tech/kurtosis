package startosis_subpackage_test

import "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"

const (
	isPartitioningEnabled = false
	emptyRunParams        = "{}"
	defaultDryRun         = false
	defaultParallelism    = 4
)

var (
	noExperimentalFeature = []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{}
)

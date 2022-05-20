package set

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
)

const clusterNameArgKey = "cluster-name"

var SetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.ClusterSetCmdStr,
	ShortDescription:         "Sets cluster to use",
	LongDescription:          "Sets the Kurtosis cluster to use based on cluster names in the Kurtosis CLI configuration file",
	RunFunc:                  run,
	Args: []*args.ArgConfig{
		{
			Key: clusterNameArgKey,
		},
	},
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	clusterName, err := args.GetNonGreedyArg(clusterNameArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to read user input.")
	}
	validClusterName, err := validateClusterName(clusterName)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to validate cluster setting.")
	}
	if !validClusterName {
		return stacktrace.NewError("Cluster name '%v' is not a valid Kurtosis cluster name.", clusterName)
	}

	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	err = clusterSettingStore.SetClusterSetting(clusterName)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to set cluster name to '%v'.", clusterName)
	}
	return nil
}

func validateClusterName(clusterName string) (bool, error) {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration when validating cluster name '%v'.", clusterName)
	}
	if _, ok := kurtosisConfig.GetKurtosisClusters()[clusterName]; ok {
		return true, nil
	}
	return false, nil
}
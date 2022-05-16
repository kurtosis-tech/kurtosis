package set

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/stacktrace"
)

const clusterNameKey = "cluster-name"

var SetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.ClusterSetCmdStr,
	ShortDescription:         "Sets the Kurtosis cluster to use.",
	LongDescription:          "Sets the Kurtosis cluster to use based on cluster names in the Kurtosis CLI configuration file.",
	RunFunc:                  run,
	Args: []*args.ArgConfig{
		{
			Key: clusterNameKey,
		},
	},
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	clusterName, err := args.GetNonGreedyArg(clusterNameKey)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to read user input from command cluster set.")
	}
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	err = clusterSettingStore.SetClusterSetting(clusterName)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to set cluster name to '%v'.", clusterName)
	}
	return nil
}
package set

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	clusterNameArgKey = "cluster-name"
)

var SetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ClusterSetCmdStr,
	ShortDescription: "Sets cluster to use",
	LongDescription:  "Sets the Kurtosis cluster to use based on cluster names in the Kurtosis CLI configuration file",
	Flags:            nil,
	Args: []*args.ArgConfig{
		{
			Key:                   clusterNameArgKey,
			IsOptional:            false,
			DefaultValue:          nil,
			IsGreedy:              false,
			ArgCompletionProvider: nil,
			ValidationFunc:        nil,
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	clusterName, err := args.GetNonGreedyArg(clusterNameArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to read user input.")
	}
	if err = validateClusterName(clusterName); err != nil {
		return stacktrace.Propagate(err, "'%s' is not a valid name for Kurtosis cluster", clusterName)
	}

	clusterUpdateSuccessful := false
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	clusterPriorToUpdate, err := clusterSettingStore.GetClusterSetting()
	if err != nil {
		return stacktrace.Propagate(err, "Tried to fetch the current cluster before changing clusters but failed")
	}

	if clusterName == clusterPriorToUpdate {
		logrus.Infof("Kurtosis cluster already set to '%s'", clusterName)
		return nil
	}

	if err = clusterSettingStore.SetClusterSetting(clusterName); err != nil {
		return stacktrace.Propagate(err, "Failed to set cluster name to '%v'.", clusterName)
	}
	defer func() {
		if clusterUpdateSuccessful {
			return
		}
		if err = clusterSettingStore.SetClusterSetting(clusterPriorToUpdate); err != nil {
			logrus.Errorf("An error happened updating cluster to '%s'. Kurtosis tried to roll back to the "+
				"previous value '%s' but the roll back failed. You have to roll back manually running "+
				"'kurtosis %s %s %s'", clusterName, clusterPriorToUpdate, command_str_consts.ClusterCmdStr,
				command_str_consts.ClusterSetCmdStr, clusterPriorToUpdate)
		}
	}()
	logrus.Infof("Cluster set to '%s', Please start a Kurtosis engine on the cluster if there isn't one already", clusterName)
	clusterUpdateSuccessful = true
	return nil
}

func validateClusterName(clusterName string) error {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration when validating cluster name '%v'.", clusterName)
	}
	if _, found := kurtosisConfig.GetKurtosisClusters()[clusterName]; !found {
		return stacktrace.NewError("Cluster '%v' isn't defined in the Kurtosis config file", clusterName)
	}
	return nil
}

package set

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	clusterNameArgKey = "cluster-name"

	noEngineVersion                        = ""
	restartEngineOnSameVersionIfAnyRunning = true
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
		logrus.Debugf("Unable to get current cluster set. If this is a fresh Kurtosis install, it's fine "+
			"as the cluster config might not be set yet. Error was: %v", err.Error())
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
			logrus.Errorf("An error happened updating cluster to '%s'. KUrtosis tried to roll back to the "+
				"previous value '%s' but the roll back failed. You have to roll back manually running "+
				"'kurtosis %s %s %s'", clusterName, clusterPriorToUpdate, command_str_consts.ClusterCmdStr,
				command_str_consts.ClusterSetCmdStr, clusterPriorToUpdate)
		}
	}()
	logrus.Infof("Cluster set to '%s', Kurtosis engine will now be restarted", clusterName)

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	// We try to do our best to restart an engine on the same version the current on is on
	_, engineClientCloseFunc, restartEngineErr := engineManager.RestartEngineIdempotently(ctx, logrus.InfoLevel, noEngineVersion, restartEngineOnSameVersionIfAnyRunning)
	if restartEngineErr != nil {
		return stacktrace.Propagate(err, "Engine could not be restarted after cluster was updated. The cluster"+
			"will be rolled back, but it is possible the engine will remain stopped. Its status can be retrieved "+
			"running 'kurtosis %s %s' and it can potentially be restarted running 'kurtosis %s %s'",
			command_str_consts.EngineCmdStr, command_str_consts.EngineStatusCmdStr, command_str_consts.EngineCmdStr,
			command_str_consts.EngineStartCmdStr)
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()
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

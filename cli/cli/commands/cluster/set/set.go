package set

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	clusterNameArgKey                     = "cluster-name"
	emptyClusterFromNeverHavingClusterSet = ""
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

	// at the moment docker clusters are local only, we stop engine in them in favor of switching to any other cluster
	stopOldEngine := false
	if clusterPriorToUpdate == emptyClusterFromNeverHavingClusterSet {
		stopOldEngine = true
	} else {
		clusterSettingPriorToUpdate, err := kurtosis_config_getter.GetKurtosisClusterConfig()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while fetching cluster setting for current cluster '%s'; not proceeding further to ensure that Kurtosis doesn't get into a bad state", clusterPriorToUpdate)
		} else if clusterSettingPriorToUpdate.GetClusterType() == resolved_config.KurtosisClusterType_Docker {
			stopOldEngine = true
		}
	}

	if stopOldEngine {
		logrus.Infof("Current cluster seems to be a local cluster of type '%s'; will stop the engine if its running so that it doesn't interfere with the updated cluster", resolved_config.KurtosisClusterType_Docker.String())
		engineManagerOldCluster, err := engine_manager.NewEngineManager(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "an error occurred while creating an engine manager for current cluster")
		}
		if err := engineManagerOldCluster.StopEngineIdempotently(ctx); err != nil {
			return stacktrace.Propagate(err, "Tried stopping engine for current cluster but failed; not proceeding further as current cluster is a local cluster and there might be port clashes with the updated cluster. Its status can be retrieved "+
				"running 'kurtosis %s %s' and it can potentially be started running 'kurtosis %s %s'",
				command_str_consts.EngineCmdStr, command_str_consts.EngineStatusCmdStr, command_str_consts.EngineCmdStr,
				command_str_consts.EngineStartCmdStr)
		}
	}

	if err = clusterSettingStore.SetClusterSetting(clusterName); err != nil {
		return stacktrace.Propagate(err, "Failed to set cluster name to '%v'.", clusterName)
	}
	logrus.Infof("Cluster set to '%s'", clusterName)
	engineManagerNewCluster, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	_, engineClientCloseFunc, err := engineManagerNewCluster.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel, defaults.DefaultEngineEnclavePoolSize)
	if err != nil {
		return stacktrace.Propagate(err, "Engine could not be started after cluster was updated. Its status can be retrieved "+
			"running 'kurtosis %s %s' and it can potentially be started running 'kurtosis %s %s'",
			command_str_consts.EngineCmdStr, command_str_consts.EngineStatusCmdStr, command_str_consts.EngineCmdStr,
			command_str_consts.EngineStartCmdStr)
	}
	defer func() {
		if err = engineClientCloseFunc(); err != nil {
			logrus.Warnf("Error closing the engine client:\n'%v'", err)
		}
	}()

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

package get

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var GetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.ClusterGetCmdStr,
	ShortDescription:         "Get current cluster",
	LongDescription:          "Get current Kurtosis cluster setting",
	RunFunc:                  run,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	clusterName, err := clusterSettingStore.GetClusterSetting()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get cluster setting.")
	}
	fmt.Fprint(logrus.StandardLogger().Out, clusterName)
	return nil
}


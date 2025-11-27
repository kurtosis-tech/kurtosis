package get

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
)

var GetCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.ClusterGetCmdStr,
	ShortDescription:         "Get current cluster",
	LongDescription:          "Get current Kurtosis cluster setting",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	clusterName, err := clusterSettingStore.GetClusterSetting()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get cluster setting.")
	}
	out.PrintOutLn(clusterName)
	return nil
}

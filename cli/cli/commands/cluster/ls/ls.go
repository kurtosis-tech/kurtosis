package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

var LsCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.ClusterLsCmdStr,
	ShortDescription:         "List valid clusters",
	LongDescription:          "List valid clusters based on defaults and the user's configuration file",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(kurtosisConfigStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration")
	}
	var clusterList []string
	for clusterName := range kurtosisConfig.GetKurtosisClusters() {
		clusterList = append(clusterList, clusterName)
	}
	sort.Strings(clusterList)
	for _, clusterName := range clusterList {
		out.PrintOutLn(clusterName)
	}
	return nil
}

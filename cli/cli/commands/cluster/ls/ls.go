package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cluster_setting"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
)

const (
	clusterCurrentColumnHeader = ""
	clusterNameColumnHeader    = "Name"

	isCurrentClusterStrIndicator = "*"
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
	clusterList, err := GetClusterList()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to get Kurtosis cluster list")
	}

	tablePrinter := output_printers.NewTablePrinter(clusterCurrentColumnHeader, clusterNameColumnHeader)

	for _, clusterName := range clusterList {
		currentClusterStr := ""
		if isCurrentCluster(clusterName) {
			currentClusterStr = isCurrentClusterStrIndicator
		}

		if err = tablePrinter.AddRow(currentClusterStr, clusterName); err != nil {
			return stacktrace.Propagate(err, "Error adding cluster to the table to be displayed")
		}
	}
	tablePrinter.Print()
	return nil
}

func GetClusterList() ([]string, error) {
	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(kurtosisConfigStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return []string{}, stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration")
	}
	var clusterList []string
	for clusterName := range kurtosisConfig.GetKurtosisClusters() {
		clusterList = append(clusterList, clusterName)
	}
	sort.Strings(clusterList)
	return clusterList, nil
}

func isCurrentCluster(clusterName string) bool {
	clusterSettingStore := kurtosis_cluster_setting.GetKurtosisClusterSettingStore()
	currentClusterName, _ := clusterSettingStore.GetClusterSetting()

	return clusterName == currentClusterName
}

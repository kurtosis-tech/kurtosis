package cluster

import (
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	get_cluster "github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/cluster/get"
	ls_cluster "github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/cluster/ls"
	set_cluster "github.com/dzobbe/PoTE-kurtosis/cli/cli/commands/cluster/set"
	"github.com/spf13/cobra"
)

// ClusterCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var ClusterCmd = &cobra.Command{
	Use:   command_str_consts.ClusterCmdStr,
	Short: "Manage Kurtosis cluster setting",
	RunE:  nil,
}

func init() {
	ClusterCmd.AddCommand(set_cluster.SetCmd.MustGetCobraCommand())
	ClusterCmd.AddCommand(ls_cluster.LsCmd.MustGetCobraCommand())
	ClusterCmd.AddCommand(get_cluster.GetCmd.MustGetCobraCommand())
}

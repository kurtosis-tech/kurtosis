package discord

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/spf13/cobra"
	"os/exec"
	"runtime"
)

const (
	mac = "darwin"
	windows = "windows"
	kurtosisDiscord = "https://discord.com/channels/783719264308953108/783719264308953111"
)

var Discord = &cobra.Command{
	Use:   command_str_consts.DiscordCmdStr,
	Short: "Opens the Kurtosis Discord",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	var subArgs []string
	switch runtime.GOOS {
		case "darwin":
			subArgs = []string{"open"}
		case  "windows":
			subArgs = []string{"cmd", "/c", "start"}
		default:
			subArgs = []string{"xdg-open"}
	}
	command := exec.Cmd{Path: subArgs[0], Args: append(args[1:], kurtosisDiscord)}
	return command.Start()
}

package discord

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
	"os/exec"
	"runtime"
)

const (
	linux = "linux"
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

func run(_ *cobra.Command, _ []string) error {
	var args []string
	switch runtime.GOOS {
		case linux:
			args = []string{"xdg-open", kurtosisDiscord}
		case mac:
			args = []string{"open", kurtosisDiscord}
		case windows:
			args = []string{"rundll32", "url.dll,FileProtocolHandler", kurtosisDiscord}
		default:
			return stacktrace.NewError("unsupported platform")
	}
	command := exec.Command(args[0], args[1:]...)
	return command.Start()
}

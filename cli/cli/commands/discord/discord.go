package discord

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
	"os/exec"
	"runtime"
)

const (
	linuxOSName = "linux"
	macOSName = "darwin"
	windowsOSName      = "windows"
	kurtosisDiscordUrl = "https://discord.com/channels/783719264308953108/783719264308953111"
)

var DiscordCmd = &cobra.Command{
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
	case linuxOSName:
		args = []string{"xdg-open", kurtosisDiscordUrl}
	case macOSName:
		args = []string{"open", kurtosisDiscordUrl}
	case windowsOSName:
		args = []string{"rundll32", "url.dll,FileProtocolHandler", kurtosisDiscordUrl}
	default:
		return stacktrace.NewError("Unsupported operating system")
	}
	command := exec.Command(args[0], args[1:]...)
	if err := command.Start(); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis DiscordCmd Channel")
	}
	return nil
}

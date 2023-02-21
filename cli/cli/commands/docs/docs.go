package docs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"os/exec"
	"runtime"
)

const (
	linuxOSName   = "linux"
	macOSName     = "darwin"
	windowsOSName = "windows"
)

var DiscordCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.DocsCmdStr,
	ShortDescription:         "Opens the Kurtosis Documentation Page",
	LongDescription:          "Opens the Kurtosis Documentation Page",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	var args []string
	switch runtime.GOOS {
	case linuxOSName:
		args = []string{"xdg-open", user_support_constants.DocumentationUrl}
	case macOSName:
		args = []string{"open", user_support_constants.DocumentationUrl}
	case windowsOSName:
		args = []string{"rundll32", "url.dll,FileProtocolHandler", user_support_constants.DocumentationUrl}
	default:
		return stacktrace.NewError("Unsupported operating system")
	}
	command := exec.Command(args[0], args[1:]...)
	if err := command.Start(); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Docs page '%v'", user_support_constants.DocumentationUrl)
	}
	return nil
}

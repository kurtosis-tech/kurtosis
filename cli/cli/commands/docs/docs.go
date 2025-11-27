package docs

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
	"os/exec"
	"runtime"
)

const (
	linuxOSName   = "linux"
	macOSName     = "darwin"
	windowsOSName = "windows"
)

var DocsCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.DocsCmdStr,
	ShortDescription:         "Opens the Kurtosis Documentation Page",
	LongDescription:          "Opens the Kurtosis Documentation Page",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

// TODO refactor this to use the multi_os_command_executor after https://github.com/dzobbe/PoTE-kurtosis/pull/28/files is merged
func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	var args []string
	switch runtime.GOOS {
	case linuxOSName:
		args = []string{"xdg-open", user_support_constants.DocumentationUrl}
	case macOSName:
		args = []string{"open", user_support_constants.DocumentationUrl}
	// TODO remove windows if we choose not to support it
	case windowsOSName:
		args = []string{"rundll32", "url.dll,FileProtocolHandler", user_support_constants.DocumentationUrl}
	default:
		return stacktrace.NewError("Unsupported operating system")
	}
	command := exec.Command(args[0], args[1:]...)
	if err := command.Start(); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the docs page '%v'", user_support_constants.DocumentationUrl)
	}
	return nil
}

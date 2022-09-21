package discord

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"os/exec"
	"runtime"
)

const (
	linuxOSName        = "linux"
	macOSName          = "darwin"
	windowsOSName      = "windows"
	kurtosisDiscordUrl = "https://discord.com/channels/783719264308953108/783719264308953111"
)

var DiscordCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:       command_str_consts.DiscordCmdStr,
	ShortDescription: "Opens the Kurtosis Discord",
	RunFunc:          run,
}

func run(
	_ context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
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

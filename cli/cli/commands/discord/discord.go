package discord

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
)

var DiscordCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.DiscordCmdStr,
	ShortDescription:         "Opens the Kurtosis Discord",
	LongDescription:          "Opens the #general channel on the Kurtosis Discord server",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	if err := multi_os_command_executor.OpenFile(user_support_constants.KurtosisDiscordUrl); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis DiscordCmd Channel")
	}
	return nil
}

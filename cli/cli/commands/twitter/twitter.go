package twitter

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/stacktrace"
)

var TwitterCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.TwitterCmdStr,
	ShortDescription:         "Opens the official Kurtosis Twitter page",
	LongDescription:          "Opens the official Kurtosis Twitter page",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	if err := multi_os_command_executor.OpenFile(user_support_constants.KurtosisTechTwitterProfileLink); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Twitter Channel")
	}
	return nil
}

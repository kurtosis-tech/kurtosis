package web

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/stacktrace"
)

var WebCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.TwitterCmdStr,
	ShortDescription:         "Opens the Kurtosis Web UI",
	LongDescription:          "Opens the Kurtosis Web UI",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

const (
	webUiLink = "http://localhost:9711"
)

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {
	if err := multi_os_command_executor.OpenFile(webUiLink); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Web UI")
	}
	return nil
}

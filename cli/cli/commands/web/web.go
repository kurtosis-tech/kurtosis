package web

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/multi_os_command_executor"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

var WebCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:               command_str_consts.WebCmdStr,
	ShortDescription:         "Opens the Kurtosis Web UI(beta)",
	LongDescription:          "Opens the Kurtosis Web UI. This feature is currently in beta.",
	Flags:                    nil,
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

const (
	webUILink = "http://localhost:9711/enclaves"
)

func run(_ context.Context, _ *flags.ParsedFlags, _ *args.ParsedArgs) error {

	contextsConfigStore := store.GetContextsConfigStore()
	currentKurtosisContext, err := contextsConfigStore.GetCurrentContext()
	if err != nil {
		return stacktrace.Propagate(err, "tried fetching the current Kurtosis context but failed, we can't switch clusters without this information. This is a bug in Kurtosis")
	}
	if store.IsRemote(currentKurtosisContext) {
		logrus.Warn("Kurtosis Cloud has been deprecated. Switch to a local kurtosis context to use the local Kurtosis web UI instead.")
	}

	if err := multi_os_command_executor.OpenFile(webUILink); err != nil {
		return stacktrace.Propagate(err, "An error occurred while opening the Kurtosis Web UI")
	}
	return nil
}

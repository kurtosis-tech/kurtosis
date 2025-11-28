package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	contexts_config_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang"
	contexts_config_generated_api "github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	contextCurrentColumnHeader = ""
	contextUuidColumnHeader    = "UUID"
	contextNameColumnHeader    = "Name"
	contextRemoteColumnHeader  = "Remote Host"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	isCurrentContextStrIndicator      = "*"
	defaultRemoteValueForLocalContext = "-"
)

var ContextLsCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ContextLsCmdStr,
	ShortDescription: "Lists Kurtosis contexts",
	LongDescription:  "Lists all the Kurtosis contexts currently configured for this installation",
	Flags: []*flags.FlagConfig{
		{
			Key:     fullUuidsFlagKey,
			Usage:   "If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
	},
	Args:                     nil,
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, flags *flags.ParsedFlags, _ *args.ParsedArgs) error {
	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	contextsConfigStore := store.GetContextsConfigStore()
	contextsConfig, err := contextsConfigStore.GetKurtosisContextsConfig()
	if err != nil {
		return stacktrace.Propagate(err, "Error retrieving currently configured contexts")
	}
	currentContextUuid := contextsConfig.GetCurrentContextUuid()

	tablePrinter := output_printers.NewTablePrinter(contextCurrentColumnHeader, contextUuidColumnHeader, contextNameColumnHeader, contextRemoteColumnHeader)
	for _, kurtosisContext := range contextsConfig.GetContexts() {
		var isCurrentIndicator string
		if kurtosisContext.GetUuid().GetValue() == currentContextUuid.GetValue() {
			isCurrentIndicator = isCurrentContextStrIndicator
		}

		var contextUuidToDisplay string
		if showFullUuids {
			contextUuidToDisplay = kurtosisContext.GetUuid().GetValue()
		} else {
			contextUuidToDisplay = uuid_generator.ShortenedUUIDString(kurtosisContext.GetUuid().GetValue())
		}

		var remoteStrToDisplay string
		contextVisitorForRemoteString := contexts_config_api.KurtosisContextVisitor[struct{}]{
			VisitLocalOnlyContextV0: func(localContext *contexts_config_generated_api.LocalOnlyContextV0) (*struct{}, error) {
				remoteStrToDisplay = defaultRemoteValueForLocalContext
				return nil, nil
			},
			VisitRemoteContextV0: func(remoteContext *contexts_config_generated_api.RemoteContextV0) (*struct{}, error) {
				remoteStrToDisplay = remoteContext.Host
				return nil, nil
			},
		}
		if _, err = contexts_config_api.Visit[struct{}](kurtosisContext, contextVisitorForRemoteString); err != nil {
			return stacktrace.Propagate(err, "Unexpected error extracting remote information from the context")
		}

		contextNameToDisplay := kurtosisContext.GetName()

		if err = tablePrinter.AddRow(isCurrentIndicator, contextUuidToDisplay, contextNameToDisplay, remoteStrToDisplay); err != nil {
			return stacktrace.Propagate(err, "Error adding context to the table to be displayed")
		}
	}
	tablePrinter.Print()
	return nil
}

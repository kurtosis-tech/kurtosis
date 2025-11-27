package add

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/api/golang/generated"
	"github.com/dzobbe/PoTE-kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
)

const (
	contextFilePathArgKey        = "context-config"
	isContextFilePathArgOptional = false
	defaultContextFilePathArg    = ""

	cloudUserIdDefaultValue     = ""
	cloudInstanceIdDefaultValue = ""
)

var ContextAddCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ContextAddCmdStr,
	ShortDescription: "Add a new Kurtosis context",
	LongDescription: "Add new Kurtosis context to the current set of configured context. Once a context is added," +
		" you can switch to it using the 'kurtosis context switch' command",
	Flags: []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		file_system_path_arg.NewFilepathArg(
			contextFilePathArgKey,
			isContextFilePathArgOptional,
			defaultContextFilePathArg,
			file_system_path_arg.DefaultValidationFunc,
		),
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	contextFilePath, err := args.GetNonGreedyArg(contextFilePathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for context file arg '%v' but none was found; "+
			"this is a bug in the Kurtosis CLI!", contextFilePathArgKey)
	}

	newContextToAdd, err := parseContextFile(contextFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to read content of context file at '%s'", contextFilePath)
	}
	return AddContext(newContextToAdd, nil, cloudInstanceIdDefaultValue, cloudUserIdDefaultValue)
}

func AddContext(newContextToAdd *generated.KurtosisContext, envVars *string, cloudInstanceId string, cloudUserId string) error {
	logrus.Infof("Adding new context '%s'", newContextToAdd.GetName())
	contextsConfigStore := store.GetContextsConfigStore()
	cloudInstanceIdCopy := new(string)
	*cloudInstanceIdCopy = cloudInstanceId
	cloudUserIdCopy := new(string)
	*cloudUserIdCopy = cloudUserId
	var enrichedContextData *generated.KurtosisContext
	if envVars != nil && *envVars != "" {
		enrichedContextData = golang.NewRemoteV0Context(
			newContextToAdd.GetUuid(),
			newContextToAdd.GetName(),
			newContextToAdd.GetRemoteContextV0().GetHost(),
			newContextToAdd.GetRemoteContextV0().GetRemotePortalPort(),
			newContextToAdd.GetRemoteContextV0().GetKurtosisBackendPort(),
			newContextToAdd.GetRemoteContextV0().GetTunnelPort(),
			newContextToAdd.GetRemoteContextV0().GetTlsConfig(),
			envVars,
			cloudUserIdCopy,
			cloudInstanceIdCopy,
		)
	} else {
		enrichedContextData = newContextToAdd
	}
	if err := contextsConfigStore.AddNewContext(enrichedContextData); err != nil {
		return stacktrace.Propagate(err, "New context '%s' with UUID '%s' could not be added to the list of "+
			"contexts already configured", enrichedContextData.GetName(), enrichedContextData.GetUuid().GetValue())
	}
	logrus.Info("Context successfully added")
	return nil
}

func parseContextFile(contextFilePath string) (*generated.KurtosisContext, error) {
	contextFileContent, err := os.ReadFile(contextFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read context of context file")
	}
	return ParseContextData(contextFileContent)
}

func ParseContextData(contextContent []byte) (*generated.KurtosisContext, error) {
	newContext := new(generated.KurtosisContext)
	if err := protojson.Unmarshal(contextContent, newContext); err != nil {
		return nil, stacktrace.Propagate(err, "Content of context file could not be parsed.")
	}
	return newContext, nil
}

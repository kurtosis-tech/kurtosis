package add

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"os"
)

const (
	contextFilePathArgKey = "context-config"
)

var ContextAddCmd = &lowlevel.LowlevelKurtosisCommand{
	CommandStr:       command_str_consts.ContextAddCmdStr,
	ShortDescription: "Add a new Kurtosis context",
	LongDescription: "Add new Kurtosis context to the current set of configured context. Once a context is added," +
		" you can switch to it using the 'kurtosis context switch' command",
	Flags: []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		{
			Key:            contextFilePathArgKey,
			ValidationFunc: validateContextFilePathArg,
		},
	},
	PreValidationAndRunFunc:  nil,
	RunFunc:                  run,
	PostValidationAndRunFunc: nil,
}

func run(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	contextFilePath, err := args.GetNonGreedyArg(contextFilePathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for context file arg '%v' but none was found; "+
			"this is a bug with Kurtosis!", contextFilePathArgKey)
	}

	contextsConfigStore := store.GetContextsConfigStore()
	newContextToAdd, err := parseContextFile(contextFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to read content of context file at '%s'", contextFilePath)
	}

	logrus.Infof("Adding new context '%s'", newContextToAdd.GetName())
	if err = contextsConfigStore.AddNewContext(newContextToAdd); err != nil {
		return stacktrace.Propagate(err, "New context '%s' with UUID '%s' could not be added to the list of "+
			"contexts already configured", newContextToAdd.GetName(), newContextToAdd.GetUuid().GetValue())
	}
	logrus.Info("Context successfully added")
	return nil
}

func parseContextFile(contextFilePath string) (*generated.KurtosisContext, error) {
	contextFileContent, err := os.ReadFile(contextFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to read context of context file")
	}

	newContext := new(generated.KurtosisContext)
	if err = protojson.Unmarshal(contextFileContent, newContext); err != nil {
		return nil, stacktrace.Propagate(err, "Content of context file at does not seem to be valid. It couldn't be parsed.")
	}
	return newContext, nil
}

func validateContextFilePathArg(_ context.Context, _ *flags.ParsedFlags, args *args.ParsedArgs) error {
	contextFilePath, err := args.GetNonGreedyArg(contextFilePathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the path to validate using key '%v'", contextFilePathArgKey)
	}
	if _, err := os.Stat(contextFilePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying context file '%v' exists and is readable",
			contextFilePath)
	}
	return nil
}

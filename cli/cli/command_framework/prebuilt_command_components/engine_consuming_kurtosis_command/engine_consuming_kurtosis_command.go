package engine_consuming_kurtosis_command

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// To avoid duplicating work, we'll instantiate the docker manager & engine client in the command's pre-validation-and-run function,
	//  and then pass it to both validation & run functions
	// This is the key where the engine client will be stored in the context
	dockerManagerCtxKey = "docker-manager"
	engineClientCtxKey = "engine-client"
	engineClientCloseFuncCtxKey = "engine-client-close-func"
)

// This is a convenience KurtosisCommand that will guarantee to start an engine
// - The Docker manager & engine connection get instantiated
// - The command takes in one or more enclave
type EngineConsumingKurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	// Will be used when displaying the command for tab completion
	ShortDescription string

	LongDescription string

	// The name of the key that will be set during PreValidationAndRun where the DockerManager can be found
	DockerManagerContextKey string

	// The name of the key that will be set during PreValidationAndRun where the engine client will be made available
	EngineClientContextKey string

	// Order isn't important here
	Flags []*flags.FlagConfig

	Args []*args.ArgConfig

	RunFunc func(
		ctx context.Context,
		dockerManager *docker_manager.DockerManager,
		engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
		flags *flags.ParsedFlags,
		args *args.ParsedArgs,
	) error
}

func (cmd *EngineConsumingKurtosisCommand) MustGetCobraCommand() *cobra.Command {
	// Validation
	if cmd.DockerManagerContextKey == engineClientCloseFuncCtxKey {
		panic(stacktrace.NewError(
			"Docker manager context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.DockerManagerContextKey,
			cmd.CommandStr,
			engineClientCloseFuncCtxKey,
		))
	}
	if cmd.EngineClientContextKey == engineClientCloseFuncCtxKey {
		panic(stacktrace.NewError(
			"Engine client context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.EngineClientContextKey,
			cmd.CommandStr,
			engineClientCloseFuncCtxKey,
		))
	}
	if cmd.DockerManagerContextKey == cmd.EngineClientContextKey {
		panic(stacktrace.NewError(
			"Docker manager context key '%v' on command '%v' is equal to engine client close function context key '%v'; this is a bug in Kurtosis!",
			cmd.DockerManagerContextKey,
			cmd.CommandStr,
			cmd.EngineClientContextKey,
		))
	}

	setupFunc := func(ctx context.Context) (context.Context, error) {
		result := ctx

		dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the Docker client")
		}
		dockerManager := docker_manager.NewDockerManager(
			logrus.StandardLogger(),
			dockerClient,
		)
		result = context.WithValue(result, cmd.DockerManagerContextKey, dockerManager)

		engineManager := engine_manager.NewEngineManager(dockerManager)
		objAttrsProvider := schema.GetObjectAttributesProvider()
		engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
		}
		result = context.WithValue(result, engineClientCtxKey, engineClient)
		result = context.WithValue(result, cmd.EngineClientContextKey, closeClientFunc)

		return result, nil
	}

	// Do the gruntwork necessary to give a Kurtosis dev the Docker manager & engine client without them
	// needing to think about how they should get it
	runFunc := func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {
		uncastedEngineClient := ctx.Value(engineClientCtxKey)
		if uncastedEngineClient == nil {
			return stacktrace.NewError("Expected an engine client to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", engineClientCtxKey)
		}
		engineClient, ok := uncastedEngineClient.(kurtosis_engine_rpc_api_bindings.EngineServiceClient)
		if !ok {
			return stacktrace.NewError("Found an object that should be the engine client stored in the context under key '%v', but this object wasn't of the correct type", engineClientCtxKey)
		}

		// TODO GET RID OF THIS!!! Everything should be doable through the engine client
		uncastedDockerManager := ctx.Value(dockerManagerCtxKey)
		if uncastedDockerManager == nil {
			return stacktrace.NewError("Expected a Docker manager to have been stored in the context under key '%v', but none was found; this is a bug in Kurtosis!", dockerManagerCtxKey)
		}
		dockerManager, ok := uncastedDockerManager.(*docker_manager.DockerManager)
		if !ok {
			return stacktrace.NewError("Found an object that should be the Docker manager stored in the context under key '%v', but this object wasn't of the correct type", dockerManagerCtxKey)
		}

		if err := cmd.RunFunc(ctx, dockerManager, engineClient, flags, args); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred calling the run function for command '%v' with short description '%v'; this is a bug in Kurtosis!",
				cmd.CommandStr,
				cmd.ShortDescription,
			 )
		}

		return nil
	}

	teardownFunc := func(ctx context.Context) {
		uncastedEngineClientCloseFunc := ctx.Value(engineClientCloseFuncCtxKey)
		if uncastedEngineClientCloseFunc != nil {
			engineClientCloseFunc, ok := uncastedEngineClientCloseFunc.(func() error)
			if ok {
				if err := engineClientCloseFunc(); err != nil {
					logrus.Warnf("We tried to close the engine client after we're done using it, but doing so threw an error:\n%v", err)
				}
			} else {
				logrus.Errorf("Expected the object at context key '%v' to be an engine client close function, but it wasn't; this is a bug in Kurtosis!", engineClientCloseFuncCtxKey)
			}
		} else {
			logrus.Errorf(
				"Expected to find an engine client close function during teardown at context key '%v', but none was found; this is a bug in Kurtosis!",
				engineClientCloseFuncCtxKey,
			)
		}
	}

	lowlevelCmd := &kurtosis_command.LowlevelKurtosisCommand{
		CommandStr:       cmd.CommandStr,
		ShortDescription: cmd.ShortDescription,
		LongDescription:  cmd.LongDescription,
		Flags:            cmd.Flags,
		Args:             cmd.Args,
		PreValidationAndRunFunc:  setupFunc,
		RunFunc:                  runFunc,
		PostValidationAndRunFunc: teardownFunc,
	}

	return lowlevelCmd.MustGetCobraCommand()
}

func getSetupFunc() {

}

func getRunFunc() {

}

func getTeardownFunc() {

}
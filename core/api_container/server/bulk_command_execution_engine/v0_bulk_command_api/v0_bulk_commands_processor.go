package v0_bulk_command_api

import (
	"context"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/kurtosis-tech/kurtosis-client/golang/bulk_command_execution/v0_bulk_command_api"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/bulk_command_execution_engine/service_ip_replacer"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type V0BulkCommandProcessor struct {
	apiService core_api_bindings.ApiContainerServiceServer
	ipReplacer *service_ip_replacer.ServiceIPReplacer
}

func NewV0BulkCommandProcessor(serviceNetwork service_network.ServiceNetwork, apiService core_api_bindings.ApiContainerServiceServer) (*V0BulkCommandProcessor, error) {
	ipReplacer, err := service_ip_replacer.NewServiceIPReplacer(
		v0_bulk_command_api.ServiceIdIpReplacementPrefix,
		v0_bulk_command_api.ServiceIdIpReplacementSuffix,
		serviceNetwork,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the v0 service ID -> IP replacer")
	}
	return &V0BulkCommandProcessor{
		apiService: apiService,
		ipReplacer: ipReplacer,
	}, nil
}

func (processor *V0BulkCommandProcessor) Process(ctx context.Context, serializedDocumentBody []byte) error {
	deserialized := new(v0_bulk_command_api.V0BulkCommands)
	if err := json.Unmarshal(serializedDocumentBody, &deserialized); err != nil {
		return stacktrace.Propagate(err, "An error occurred deserializing the bulk commands object")
	}

	for idx, command := range deserialized.Commands {
		if err := processor.parseAndExecuteCommand(ctx, command); err != nil {
			return stacktrace.Propagate(err, "An error occurred parsing and executing command #%v", idx)
		}
	}
	return nil
}

func (processor *V0BulkCommandProcessor) parseAndExecuteCommand(ctx context.Context, command v0_bulk_command_api.V0SerializableCommand) error {
	switch command.Type {
	case v0_bulk_command_api.RegisterServiceCommandType:
		args, ok := command.Args.(*core_api_bindings.RegisterServiceArgs)
		if !ok {
			return stacktrace.NewError("An error occurred downcasting the generic args object to register service args")
		}
		if _, err := processor.apiService.RegisterService(ctx, args); err != nil {
			return stacktrace.Propagate(err, "An error occurred processing the register service command")
		}
	case v0_bulk_command_api.StartServiceCommandType:
		args, ok := command.Args.(*core_api_bindings.StartServiceArgs)
		if !ok {
			return stacktrace.NewError("An error occurred downcasting the generic args object to start service args")
		}
		ipReplacedArgs, err := processor.doServiceIdToIpReplacementOnStartServiceArgs(args)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred executing service ID -> IP replacement on the start service args")
		}
		hostPortBindings, err := processor.apiService.StartService(ctx, ipReplacedArgs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred processing the start service command")
		}
		logrus.Infof("Started service '%v' via bulk command, resulting in the following host port bindings: %+v", ipReplacedArgs.ServiceId, hostPortBindings)
	case v0_bulk_command_api.RemoveServiceCommandType:
		args, ok := command.Args.(*core_api_bindings.RemoveServiceArgs)
		if !ok {
			return stacktrace.NewError("An error occurred downcasting the generic args object to remove service args")
		}
		if _, err := processor.apiService.RemoveService(ctx, args); err != nil {
			return stacktrace.Propagate(err, "An error occurred processing the remove service command")
		}
	// TODO MORE COMMANDS
	default:
		return stacktrace.NewError("Unrecognized bulk command type '%v'", command.Type)
	}

	return nil
}

func (processor *V0BulkCommandProcessor) doServiceIdToIpReplacementOnStartServiceArgs(args *core_api_bindings.StartServiceArgs) (*core_api_bindings.StartServiceArgs, error) {
	ipReplacedArgs := proto.Clone(args).(*core_api_bindings.StartServiceArgs)

	replacedEntrypointArgs, err := processor.ipReplacer.ReplaceStrSlice(args.EntrypointArgs)
	ipReplacedArgs.EntrypointArgs = replacedEntrypointArgs

	replacedCmdArgs, err := processor.ipReplacer.ReplaceStrSlice(args.CmdArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the CMD arguments")
	}
	ipReplacedArgs.CmdArgs = replacedCmdArgs

	replacedEnvVars, err := processor.ipReplacer.ReplaceMapValues(args.DockerEnvVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the env var values")
	}
	ipReplacedArgs.DockerEnvVars = replacedEnvVars

	replacedFilesArtifactMountDirpaths, err := processor.ipReplacer.ReplaceMapValues(args.FilesArtifactMountDirpaths)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the files artifact mount dirpaths")
	}
	ipReplacedArgs.FilesArtifactMountDirpaths = replacedFilesArtifactMountDirpaths

	replacedSuiteExVolMntDirpath, err := processor.ipReplacer.ReplaceStr(args.SuiteExecutionVolMntDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the suite execution volume mount dirpath")
	}
	ipReplacedArgs.SuiteExecutionVolMntDirpath = replacedSuiteExVolMntDirpath

	return ipReplacedArgs, nil
}

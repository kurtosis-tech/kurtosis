/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package v0_bulk_command_api

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-client/golang/core_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/bulk_command_execution_engine/service_ip_replacer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	successExecCmdExitCode = 0
)

type v0CommandProcessingVisitor struct {
	// Normally storing a context in an object is prohibited, but this is a special case as the visitor acts more like a transient function than a long-lived struct
	ctx                    context.Context
	uncastedCommandArgsPtr proto.Message // POINTER to the arg object!
	ipReplacer             *service_ip_replacer.ServiceIPReplacer
	apiService             core_api_bindings.ApiContainerServiceServer
}

func newV0CommandProcessingVisitor(ctx context.Context, uncastedCommandArgsPtr proto.Message, ipReplacer *service_ip_replacer.ServiceIPReplacer, apiService core_api_bindings.ApiContainerServiceServer) *v0CommandProcessingVisitor {
	return &v0CommandProcessingVisitor{ctx: ctx, uncastedCommandArgsPtr: uncastedCommandArgsPtr, ipReplacer: ipReplacer, apiService: apiService}
}


// ====================================================================================================
//                                         Public functions
// ====================================================================================================

func (visitor v0CommandProcessingVisitor) VisitRegisterService() error {
	logrus.Infof("Uncasted: %+v", visitor.uncastedCommandArgsPtr)
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.RegisterServiceArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to register service args")
	}
	if _, err := visitor.apiService.RegisterService(visitor.ctx, castedArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred registering service with ID '%v'", castedArgs.ServiceId)
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitGenerateFiles() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.GenerateFilesArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to generate files args")
	}
	if _, err := visitor.apiService.GenerateFiles(visitor.ctx, castedArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred generating the requested files")
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitStartService() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.StartServiceArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to start service args")
	}
	ipReplacedArgs, err := visitor.doServiceIdToIpReplacementOnStartServiceArgs(castedArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing service ID -> IP replacement on the start service args")
	}
	hostPortBindings, err := visitor.apiService.StartService(visitor.ctx, ipReplacedArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting service '%v'", castedArgs.ServiceId)
	}
	logrus.Infof("Started service '%v' via bulk command, resulting in the following host port bindings: %+v", ipReplacedArgs.ServiceId, hostPortBindings)
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitRemoveService() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.RemoveServiceArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to remove service args")
	}
	if _, err := visitor.apiService.RemoveService(visitor.ctx, castedArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v'", castedArgs.ServiceId)
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitRepartition() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.RepartitionArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to repartition args")
	}
	if _, err := visitor.apiService.Repartition(visitor.ctx, castedArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitionoing the network")
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitExecCommand() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.ExecCommandArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to exec command args")
	}
	replacedArgs, err := visitor.doServiceIdToIpReplacementOnExecCommandArgs(castedArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred doing service ID -> IP replacement on the exec command args")
	}
	execCmdResponse, err := visitor.apiService.ExecCommand(visitor.ctx, replacedArgs)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred running exec command '%+v' on service '%v'",
			replacedArgs.CommandArgs,
			replacedArgs.ServiceId,
		)
	}
	if execCmdResponse.ExitCode != successExecCmdExitCode {
		return stacktrace.NewError(
			"Exec command '%+v' on service '%v' exited with non-%v exit code '%v'",
			replacedArgs.CommandArgs,
			replacedArgs.ServiceId,
			successExecCmdExitCode,
			execCmdResponse.ExitCode,
		)
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitWaitForEndpointAvailability() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.WaitForEndpointAvailabilityArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to repartition args")
	}
	replacedArgs, err := visitor.doServiceIdToIpReplacementOnWaitForEndpointAvailabilityArgs(castedArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred doing service ID -> IP replacement on the endpoint availability-waiting args")
	}
	if _, err := visitor.apiService.WaitForEndpointAvailability(visitor.ctx, replacedArgs); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred waiting for availability of endpoint at path '%v' on service '%v' at port '%v'",
			replacedArgs.Path,
			replacedArgs.ServiceId,
			replacedArgs.Port,
		)
	}
	return nil
}

func (visitor v0CommandProcessingVisitor) VisitExecuteBulkCommands() error {
	castedArgs, ok := visitor.uncastedCommandArgsPtr.(*core_api_bindings.ExecuteBulkCommandsArgs)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting the generic args object to bulk command execution args")
	}
	if _, err := visitor.apiService.ExecuteBulkCommands(visitor.ctx, castedArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the bulk commands")
	}
	return nil
}



// ====================================================================================================
//                                     Private helper functions
// ====================================================================================================

func (visitor v0CommandProcessingVisitor) doServiceIdToIpReplacementOnWaitForEndpointAvailabilityArgs(
		args *core_api_bindings.WaitForEndpointAvailabilityArgs) (*core_api_bindings.WaitForEndpointAvailabilityArgs, error) {
	clonedMessage := proto.Clone(args)
	ipReplacedArgs, ok := clonedMessage.(*core_api_bindings.WaitForEndpointAvailabilityArgs)
	if !ok {
		return nil, stacktrace.NewError("Couldn't downcast the cloned proto message to endpoint availability-waiting args")
	}

	replacedPath, err := visitor.ipReplacer.ReplaceStr(args.Path)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing service ID -> IP replacement on path '%v'", args.Path)
	}
	ipReplacedArgs.Path = replacedPath

	replacedBodyText, err := visitor.ipReplacer.ReplaceStr(args.BodyText)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred doing service ID -> IP replacement on body text '%v'", args.BodyText)
	}
	ipReplacedArgs.BodyText = replacedBodyText

	return ipReplacedArgs, nil
}

func (visitor v0CommandProcessingVisitor) doServiceIdToIpReplacementOnExecCommandArgs(args *core_api_bindings.ExecCommandArgs) (*core_api_bindings.ExecCommandArgs, error) {
	clonedMessage := proto.Clone(args)
	ipReplacedArgs, ok := clonedMessage.(*core_api_bindings.ExecCommandArgs)
	if !ok {
		return nil, stacktrace.NewError("Couldn't downcast the cloned proto message to start service args")
	}

	replacedCmd, err := visitor.ipReplacer.ReplaceStrSlice(args.CommandArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing service ID -> IP replacement on command to exec '%+v'", args.CommandArgs)
	}
	ipReplacedArgs.CommandArgs = replacedCmd

	return ipReplacedArgs, nil
}

func (visitor v0CommandProcessingVisitor) doServiceIdToIpReplacementOnStartServiceArgs(args *core_api_bindings.StartServiceArgs) (*core_api_bindings.StartServiceArgs, error) {
	clonedMessage := proto.Clone(args)
	ipReplacedArgs, ok := clonedMessage.(*core_api_bindings.StartServiceArgs)
	if !ok {
		return nil, stacktrace.NewError("Couldn't downcast the cloned proto message to start service args")
	}

	replacedEntrypointArgs, err := visitor.ipReplacer.ReplaceStrSlice(args.EntrypointArgs)
	ipReplacedArgs.EntrypointArgs = replacedEntrypointArgs

	replacedCmdArgs, err := visitor.ipReplacer.ReplaceStrSlice(args.CmdArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the CMD arguments")
	}
	ipReplacedArgs.CmdArgs = replacedCmdArgs

	replacedEnvVars, err := visitor.ipReplacer.ReplaceMapValues(args.DockerEnvVars)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the env var values")
	}
	ipReplacedArgs.DockerEnvVars = replacedEnvVars

	replacedFilesArtifactMountDirpaths, err := visitor.ipReplacer.ReplaceMapValues(args.FilesArtifactMountDirpaths)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the files artifact mount dirpaths")
	}
	ipReplacedArgs.FilesArtifactMountDirpaths = replacedFilesArtifactMountDirpaths

	replacedSuiteExVolMntDirpath, err := visitor.ipReplacer.ReplaceStr(args.SuiteExecutionVolMntDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing service IDs with IPs for the suite execution volume mount dirpath")
	}
	ipReplacedArgs.SuiteExecutionVolMntDirpath = replacedSuiteExVolMntDirpath

	return ipReplacedArgs, nil
}

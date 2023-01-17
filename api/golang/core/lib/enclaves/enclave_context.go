/*
 *    Copyright 2021 Kurtosis Technologies Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */

package enclaves

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"math"
	"path"
)

type EnclaveUUID string

type PartitionID string

const (
	kurtosisYamlFilename                      = "kurtosis.yml"
	ensureCompressedFileIsLesserThanGRPCLimit = true
)

// Docs available at https://docs.kurtosis.com/sdk/#enclavecontext
type EnclaveContext struct {
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	enclaveUuid EnclaveUUID
	enclaveName string
}

/*
Creates a new EnclaveContext object with the given parameters.
*/
// TODO Migrate this to take in API container IP & API container GRPC port num, to match the (better) way that
//  Typescript does it, so that the user doesn't have to figure out how to instantiate the ApiContainerServiceClient on their own!
func NewEnclaveContext(
	client kurtosis_core_rpc_api_bindings.ApiContainerServiceClient,
	enclaveUuid EnclaveUUID,
	enclaveName string,
) *EnclaveContext {
	return &EnclaveContext{
		client:      client,
		enclaveUuid: enclaveUuid,
		enclaveName: enclaveName,
	}
}

// Docs available at https://docs.kurtosis.com/sdk/#getenclaveuuid---enclaveuuid
func (enclaveCtx *EnclaveContext) GetEnclaveUuid() EnclaveUUID {
	return enclaveCtx.enclaveUuid
}

// Docs available at https://docs.kurtosis.com/sdk/#getenclavename---string
func (enclaveCtx *EnclaveContext) GetEnclaveName() string {
	return enclaveCtx.enclaveName
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptstring-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkScript(ctx context.Context, serializedScript string, serializedParams string, dryRun bool) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	executeStartosisScriptArgs := binding_constructors.NewRunStarlarkScriptArgs(serializedScript, serializedParams, dryRun)
	starlarkResponseLineChan := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)

	stream, err := enclaveCtx.client.RunStarlarkScript(ctxWithCancel, executeStartosisScriptArgs)
	if err != nil {
		cancelCtxFunc() // manually call the cancel function as something went wrong
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened executing Kurtosis script.")
	}

	go runReceiveStarlarkResponseLineRoutine(cancelCtxFunc, stream, starlarkResponseLineChan)
	return starlarkResponseLineChan, cancelCtxFunc, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptblockingstring-serializedstarlarkscript-boolean-dryrun---starlarkrunresult-runresult-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkScriptBlocking(ctx context.Context, serializedScript string, serializedParams string, dryRun bool) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkScript(ctx, serializedScript, serializedParams, dryRun)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error running Starlark Script")
	}
	return ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan), nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackagestring-packagerootpath-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkPackage(ctx context.Context, packageRootPath string, serializedParams string, dryRun bool) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	starlarkResponseLineChan := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	executeStartosisPackageArgs, err := enclaveCtx.assembleRunStartosisPackageArg(packageRootPath, serializedParams, dryRun)
	if err != nil {
		cancelCtxFunc() // manually call the cancel function as something went wrong
		return nil, nil, stacktrace.Propagate(err, "Error preparing package for execution '%v'", packageRootPath)
	}

	stream, err := enclaveCtx.client.RunStarlarkPackage(ctxWithCancel, executeStartosisPackageArgs)
	if err != nil {
		cancelCtxFunc() // manually call the cancel function as something went wrong
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened executing Starlark package '%v'", packageRootPath)
	}

	go runReceiveStarlarkResponseLineRoutine(cancelCtxFunc, stream, starlarkResponseLineChan)
	return starlarkResponseLineChan, cancelCtxFunc, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackageblockingstring-packagerootpath-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkPackageBlocking(ctx context.Context, packageRootPath string, serializedParams string, dryRun bool) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkPackage(ctx, packageRootPath, serializedParams, dryRun)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error running Starlark package")
	}
	return ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan), nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackagestring-packageid-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkRemotePackage(ctx context.Context, packageId string, serializedParams string, dryRun bool) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	starlarkResponseLineChan := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	executeStartosisScriptArgs := binding_constructors.NewRunStarlarkRemotePackageArgs(packageId, serializedParams, dryRun)

	stream, err := enclaveCtx.client.RunStarlarkPackage(ctxWithCancel, executeStartosisScriptArgs)
	if err != nil {
		cancelCtxFunc() // manually call the cancel function as something went wrong
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened executing Starlark package '%v'", packageId)
	}

	go runReceiveStarlarkResponseLineRoutine(cancelCtxFunc, stream, starlarkResponseLineChan)
	return starlarkResponseLineChan, cancelCtxFunc, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackageblockingstring-packageid-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkRemotePackageBlocking(ctx context.Context, packageId string, serializedParams string, dryRun bool) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkRemotePackage(ctx, packageId, serializedParams, dryRun)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error running remote Starlark package")
	}
	return ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan), nil
}

// Docs available at https://docs.kurtosis.com/sdk/#getservicecontextserviceid-serviceid---servicecontext-servicecontext
func (enclaveCtx *EnclaveContext) GetServiceContext(serviceId services.ServiceID) (*services.ServiceContext, error) {
	serviceIdMapForArgs := map[string]bool{string(serviceId): true}
	getServiceInfoArgs := binding_constructors.NewGetServicesArgs(serviceIdMapForArgs)
	response, err := enclaveCtx.client.GetServices(context.Background(), getServiceInfoArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get info for service '%v'",
			serviceId)
	}
	serviceInfo, found := response.GetServiceInfo()[string(serviceId)]
	if !found {
		return nil, stacktrace.NewError("Failed to retrieve service information for service '%v'", string(serviceId))
	}
	if serviceInfo.GetPrivateIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty private IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceId)
	}

	serviceCtxPrivatePorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetPrivatePorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the private ports returned by the API to ports usable by the service context")
	}
	serviceCtxPublicPorts, err := convertApiPortsToServiceContextPorts(serviceInfo.GetMaybePublicPorts())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the public ports returned by the API to ports usable by the service context")
	}

	serviceContext := services.NewServiceContext(
		enclaveCtx.client,
		serviceId,
		services.ServiceGUID(serviceInfo.ServiceGuid),
		serviceInfo.GetPrivateIpAddr(),
		serviceCtxPrivatePorts,
		serviceInfo.GetMaybePublicIpAddr(),
		serviceCtxPublicPorts,
	)

	return serviceContext, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#getservices---mapserviceid--serviceguid-serviceids
func (enclaveCtx *EnclaveContext) GetServices() (map[services.ServiceID]services.ServiceGUID, error) {
	getServicesArgs := binding_constructors.NewGetServicesArgs(map[string]bool{})
	response, err := enclaveCtx.client.GetServices(context.Background(), getServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the service IDs in the enclave")
	}

	serviceInfos := make(map[services.ServiceID]services.ServiceGUID, len(response.GetServiceInfo()))
	for serviceIdStr, responseServiceInfo := range response.GetServiceInfo() {
		serviceId := services.ServiceID(serviceIdStr)
		serviceGuid := services.ServiceGUID(responseServiceInfo.GetServiceGuid())
		serviceInfos[serviceId] = serviceGuid
	}
	return serviceInfos, nil
}

// Docs available at https://docs.kurtosis.com/sdk#uploadfilesstring-pathtoupload-string-artifactname
func (enclaveCtx *EnclaveContext) UploadFiles(pathToUpload string, artifactName string) (services.FilesArtifactUUID, error) {
	content, err := shared_utils.CompressPath(pathToUpload, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return "", stacktrace.Propagate(err,
			"There was an error compressing the file '%v' before upload",
			pathToUpload)
	}

	args := binding_constructors.NewUploadFilesArtifactArgs(content, artifactName)
	response, err := enclaveCtx.client.UploadFilesArtifact(context.Background(), args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error was encountered while uploading data to the API Container.")
	}
	return services.FilesArtifactUUID(response.Uuid), nil
}

// Docs available at https://docs.kurtosis.com/sdk#storewebfilesstring-urltodownload-string-artifactname
func (enclaveCtx *EnclaveContext) StoreWebFiles(ctx context.Context, urlToStoreWeb string, artifactName string) (services.FilesArtifactUUID, error) {
	args := binding_constructors.NewStoreWebFilesArtifactArgs(urlToStoreWeb, artifactName)
	response, err := enclaveCtx.client.StoreWebFilesArtifact(ctx, args)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred downloading files artifact from URL '%v'", urlToStoreWeb)
	}
	return services.FilesArtifactUUID(response.Uuid), nil
}

// Docs available at https://docs.kurtosis.com/sdk#downloadfilesartifact-fileidentifier-string
func (enclaveCtx *EnclaveContext) DownloadFilesArtifact(ctx context.Context, artifactIdentifier string) ([]byte, error) {
	args := binding_constructors.DownloadFilesArtifactArgs(artifactIdentifier)
	response, err := enclaveCtx.client.DownloadFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred downloading files artifact '%v'", artifactIdentifier)
	}
	return response.Data, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func convertApiPortsToServiceContextPorts(apiPorts map[string]*kurtosis_core_rpc_api_bindings.Port) (map[string]*services.PortSpec, error) {
	result := map[string]*services.PortSpec{}
	for portId, apiPortSpec := range apiPorts {
		apiTransportProtocol := apiPortSpec.GetTransportProtocol()
		serviceCtxTransportProtocol := services.TransportProtocol(apiTransportProtocol)
		if !serviceCtxTransportProtocol.IsValid() {
			return nil, stacktrace.NewError("Received unrecognized protocol '%v' from the API", apiTransportProtocol)
		}
		portNumUint32 := apiPortSpec.GetNumber()
		if portNumUint32 > math.MaxUint16 {
			return nil, stacktrace.NewError(
				"Received port num '%v' from the API which is higher than the max uint16 value '%v'; this is VERY weird because ports should be 16-bit numbers",
				portNumUint32,
				math.MaxUint16,
			)
		}

		portNumUint16 := uint16(portNumUint32)
		apiMaybeApplicationProtocol := apiPortSpec.GetMaybeApplicationProtocol()
		result[portId] = services.NewPortSpec(
			portNumUint16,
			serviceCtxTransportProtocol,
			apiMaybeApplicationProtocol,
		)
	}
	return result, nil
}

func runReceiveStarlarkResponseLineRoutine(cancelCtxFunc context.CancelFunc, stream grpc.ClientStream, kurtosisResponseLineChan chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) {
	defer func() {
		close(kurtosisResponseLineChan)
		cancelCtxFunc()
	}()
	for {
		responseLine := new(kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
		err := stream.RecvMsg(responseLine)
		if err == io.EOF {
			logrus.Debugf("Successfully reached the end of the response stream. Closing.")
			return
		}
		if err != nil {
			logrus.Errorf("Unexpected error happened reading the stream. Client might have cancelled the stream\n%v", err.Error())
			return
		}
		kurtosisResponseLineChan <- responseLine
	}
}

func (enclaveCtx *EnclaveContext) assembleRunStartosisPackageArg(packageRootPath string, serializedParams string, dryRun bool) (*kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, error) {
	kurtosisYamlFilepath := path.Join(packageRootPath, kurtosisYamlFilename)

	kurtosisYaml, err := parseKurtosisYaml(kurtosisYamlFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error parsing the '%v' at '%v'", kurtosisYamlFilename, packageRootPath)
	}

	logrus.Infof("Compressing package '%v' at '%v' for upload", kurtosisYaml.PackageName, packageRootPath)
	compressedModule, err := shared_utils.CompressPath(packageRootPath, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error compressing module '%v' before upload", packageRootPath)
	}
	logrus.Infof("Uploading and executing package '%v'", kurtosisYaml.PackageName)
	return binding_constructors.NewRunStarlarkPackageArgs(kurtosisYaml.PackageName, compressedModule, serializedParams, dryRun), nil
}

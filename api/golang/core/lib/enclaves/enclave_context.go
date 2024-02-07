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
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"

	yaml_convert "github.com/ghodss/yaml"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang/grpc_file_streaming"
	"github.com/kurtosis-tech/kurtosis/utils"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EnclaveUUID string

const (
	kurtosisYamlFilename    = "kurtosis.yml"
	enforceMaxFileSizeLimit = true

	osPathSeparatorString = string(os.PathSeparator)

	dotRelativePathIndicatorString = "."

	// required to get around "only Github URLs" validation
	composePackageIdPlaceholder = "github.com/NOTIONAL_USER/USER_UPLOADED_COMPOSE_PACKAGE"
)

// TODO Remove this once package ID is detected ONLY the APIC side (i.e. the CLI doesn't need to tell the APIC what package ID it's using)
// Doing so requires that we upload completely anonymous packages to the APIC, and it figures things out from there
var supportedDockerComposeYmlFilenames = []string{
	"compose.yml",
	"compose.yaml",
	"docker-compose.yml",
	"docker-compose.yaml",
	"docker_compose.yml",
	"docker_compose.yaml",
}

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
func (enclaveCtx *EnclaveContext) RunStarlarkScript(
	ctx context.Context,
	serializedScript string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	oldSerializedParams := runConfig.SerializedParams
	serializedParams, err := maybeParseYaml(oldSerializedParams)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred when parsing YAML args for script '%v'", oldSerializedParams)
	}
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	executeStartosisScriptArgs := binding_constructors.NewRunStarlarkScriptArgs(runConfig.MainFunctionName, serializedScript, serializedParams, runConfig.DryRun, runConfig.Parallelism, runConfig.ExperimentalFeatureFlags, runConfig.CloudInstanceId, runConfig.CloudUserId, runConfig.ImageDownload)
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
func (enclaveCtx *EnclaveContext) RunStarlarkScriptBlocking(
	ctx context.Context,
	serializedScript string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkScript(ctx, serializedScript, runConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error requesting Starlark Script run")
	}
	starlarkResponse := ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan)
	return starlarkResponse, getErrFromStarlarkRunResult(starlarkResponse)
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackagestring-packagerootpath-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkPackage(
	ctx context.Context,
	packageRootPath string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	serializedParams, err := maybeParseYaml(runConfig.SerializedParams)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred when parsing YAML args for package '%v'", serializedParams)
	}
	executionStartedSuccessfully := false
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	defer func() {
		if !executionStartedSuccessfully {
			cancelCtxFunc() // cancel the context as something went wrong
		}
	}()

	starlarkResponseLineChan := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)

	packageName, packageReplaceOptions, err := getPackageNameAndReplaceOptions(packageRootPath)
	if err != nil {
		return nil, nil, err
	}

	executeStartosisPackageArgs, err := enclaveCtx.assembleRunStartosisPackageArg(
		packageName,
		runConfig.RelativePathToMainFile,
		runConfig.MainFunctionName,
		serializedParams,
		runConfig.DryRun,
		runConfig.Parallelism,
		runConfig.ExperimentalFeatureFlags,
		runConfig.CloudInstanceId,
		runConfig.CloudUserId,
		runConfig.ImageDownload)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Error preparing package '%s' for execution", packageRootPath)
	}

	err = enclaveCtx.uploadStarlarkPackage(executeStartosisPackageArgs.PackageId, packageRootPath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Error uploading package '%s' prior to executing it", packageRootPath)
	}

	if len(packageReplaceOptions) > 0 {
		if err = enclaveCtx.uploadLocalStarlarkPackageDependencies(packageRootPath, packageReplaceOptions); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred while uploading the local starlark package dependencies from the replace options '%+v'", packageReplaceOptions)
		}
	}

	stream, err := enclaveCtx.client.RunStarlarkPackage(ctxWithCancel, executeStartosisPackageArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened executing Starlark package '%v'", packageRootPath)
	}

	go runReceiveStarlarkResponseLineRoutine(cancelCtxFunc, stream, starlarkResponseLineChan)
	executionStartedSuccessfully = true
	return starlarkResponseLineChan, cancelCtxFunc, nil
}

// Determines the package name and replace options based on [packageRootPath]
// If a kurtosis.yml is detected, package is a kurtosis package
// If a valid [supportedDockerComposeYaml] is detected, package is a docker compose package
func getPackageNameAndReplaceOptions(packageRootPath string) (string, map[string]string, error) {
	var packageName string
	var packageReplaceOptions map[string]string

	// use kurtosis package if it exists
	if _, err := os.Stat(path.Join(packageRootPath, kurtosisYamlFilename)); err == nil {
		kurtosisYml, err := getKurtosisYaml(packageRootPath)
		if err != nil {
			return "", map[string]string{}, stacktrace.Propagate(err, "An error occurred getting Kurtosis yaml file from path '%s'", packageRootPath)
		}
		packageName = kurtosisYml.PackageName
		packageReplaceOptions = kurtosisYml.PackageReplaceOptions
	} else {
		// use compose package if it exists
		composeAbsFilepath := ""
		for _, candidateComposeFilename := range supportedDockerComposeYmlFilenames {
			candidateComposeAbsFilepath := path.Join(packageRootPath, candidateComposeFilename)
			if _, err := os.Stat(candidateComposeAbsFilepath); err == nil {
				composeAbsFilepath = candidateComposeAbsFilepath
				break
			}
		}
		if composeAbsFilepath == "" {
			return "", map[string]string{}, stacktrace.NewError(
				"Neither a '%s' file nor one of the default Compose files (%s) was found in the package root; at least one of these is required",
				kurtosisYamlFilename,
				strings.Join(supportedDockerComposeYmlFilenames, ", "),
			)
		}
		packageName = composePackageIdPlaceholder
		packageReplaceOptions = map[string]string{}
	}

	return packageName, packageReplaceOptions, nil
}

func (enclaveCtx *EnclaveContext) uploadLocalStarlarkPackageDependencies(packageRootPath string, packageReplaceOptions map[string]string) error {
	for dependencyPackageId, replaceOption := range packageReplaceOptions {
		if isLocalDependencyReplace(replaceOption) {
			localPackagePath := path.Join(packageRootPath, replaceOption)
			if err := enclaveCtx.uploadStarlarkPackage(dependencyPackageId, localPackagePath); err != nil {
				return stacktrace.Propagate(err, "Error uploading package '%s' prior to executing it", replaceOption)
			}
		}
	}
	return nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackageblockingstring-packagerootpath-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkPackageBlocking(
	ctx context.Context,
	packageRootPath string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkPackage(ctx, packageRootPath, runConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error running Starlark package")
	}
	starlarkResponse := ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan)
	return starlarkResponse, getErrFromStarlarkRunResult(starlarkResponse)
}

func maybeParseYaml(serializedParams string) (string, error) {
	if valid := isValidJSON(serializedParams); valid {
		return serializedParams, nil
	}
	logrus.Debugf("Serialized params '%v' is not valid JSON, trying to convert from YAML", serializedParams)
	var err error
	serializedParamsBytes, err := yaml_convert.YAMLToJSON([]byte(serializedParams))
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed while converting serialized params to json")
	}
	serializedParamsStr := string(serializedParamsBytes)
	logrus.Debugf("Converted to '%v'", serializedParamsStr)
	return serializedParamsStr, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackagestring-packageid-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkRemotePackage(
	ctx context.Context,
	packageId string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
	serializedParams, err := maybeParseYaml(runConfig.SerializedParams)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred when parsing YAML args for remote package '%v'", serializedParams)
	}
	executionStartedSuccessfully := false
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	defer func() {
		if !executionStartedSuccessfully {
			cancelCtxFunc() // cancel the context as something went wrong
		}
	}()

	starlarkResponseLineChan := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	executeStartosisScriptArgs := binding_constructors.NewRunStarlarkRemotePackageArgs(packageId, runConfig.RelativePathToMainFile, runConfig.MainFunctionName, serializedParams, runConfig.DryRun, runConfig.Parallelism, runConfig.ExperimentalFeatureFlags, runConfig.CloudInstanceId, runConfig.CloudUserId, runConfig.ImageDownload)

	stream, err := enclaveCtx.client.RunStarlarkPackage(ctxWithCancel, executeStartosisScriptArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Unexpected error happened executing Starlark package '%v'", packageId)
	}

	go runReceiveStarlarkResponseLineRoutine(cancelCtxFunc, stream, starlarkResponseLineChan)
	executionStartedSuccessfully = true
	return starlarkResponseLineChan, cancelCtxFunc, nil
}

func isValidJSON(maybeJson string) bool {
	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(maybeJson), &jsonObj); err != nil {
		return false
	}
	logrus.Debugf("Valid json found '%v'", jsonObj)
	return true
}

// Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackageblockingstring-packageid-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
func (enclaveCtx *EnclaveContext) RunStarlarkRemotePackageBlocking(
	ctx context.Context,
	packageId string,
	runConfig *starlark_run_config.StarlarkRunConfig,
) (*StarlarkRunResult, error) {
	starlarkRunResponseLineChan, _, err := enclaveCtx.RunStarlarkRemotePackage(ctx, packageId, runConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error running remote Starlark package")
	}
	starlarkResponse := ReadStarlarkRunResponseLineBlocking(starlarkRunResponseLineChan)
	return starlarkResponse, getErrFromStarlarkRunResult(starlarkResponse)
}

// Docs available at https://docs.kurtosis.com/sdk#getservicecontextstring-serviceidentifier---servicecontext-servicecontext
func (enclaveCtx *EnclaveContext) GetServiceContext(serviceIdentifier string) (*services.ServiceContext, error) {
	serviceIdentifierMapForArgs := map[string]bool{serviceIdentifier: true}
	getServiceInfoArgs := binding_constructors.NewGetServicesArgs(serviceIdentifierMapForArgs)
	response, err := enclaveCtx.client.GetServices(context.Background(), getServiceInfoArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get info for service '%v'",
			serviceIdentifier)
	}
	serviceInfo, found := response.GetServiceInfo()[serviceIdentifier]
	if !found {
		return nil, stacktrace.NewError("Failed to retrieve service information for service '%v'", serviceIdentifier)
	}
	if serviceInfo.GetPrivateIpAddr() == "" {
		return nil, stacktrace.NewError(
			"Kurtosis API reported an empty private IP address for service '%v' - this should never happen, and is a bug with Kurtosis!",
			serviceIdentifier)
	}

	serviceContext, err := enclaveCtx.convertServiceInfoToServiceContext(serviceInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the service info to a service context.")
	}

	return serviceContext, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getservicecontexts--mapstring--bool---servicecontext-servicecontext
func (enclaveCtx *EnclaveContext) GetServiceContexts(serviceIdentifiers map[string]bool) (map[services.ServiceName]*services.ServiceContext, error) {
	getServiceInfoArgs := binding_constructors.NewGetServicesArgs(serviceIdentifiers)
	response, err := enclaveCtx.client.GetServices(context.Background(), getServiceInfoArgs)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get info for services '%v'",
			serviceIdentifiers)
	}

	serviceContexts := make(map[services.ServiceName]*services.ServiceContext, len(response.GetServiceInfo()))
	for _, serviceInfo := range response.GetServiceInfo() {
		serviceContext, err := enclaveCtx.convertServiceInfoToServiceContext(serviceInfo)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting the service info to a service context.")
		}
		serviceContexts[serviceContext.GetServiceName()] = serviceContext
	}

	return serviceContexts, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getservices---mapservicename--serviceuuid-serviceidentifiers
func (enclaveCtx *EnclaveContext) GetServices() (map[services.ServiceName]services.ServiceUUID, error) {
	getServicesArgs := binding_constructors.NewGetServicesArgs(map[string]bool{})
	response, err := enclaveCtx.client.GetServices(context.Background(), getServicesArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the services in the enclave")
	}

	serviceInfos := make(map[services.ServiceName]services.ServiceUUID, len(response.GetServiceInfo()))
	for serviceIdStr, responseServiceInfo := range response.GetServiceInfo() {
		serviceName := services.ServiceName(serviceIdStr)
		serviceUuid := services.ServiceUUID(responseServiceInfo.GetServiceUuid())
		serviceInfos[serviceName] = serviceUuid
	}
	return serviceInfos, nil
}

// Docs available at https://docs.kurtosis.com/sdk#uploadfilesstring-pathtoupload-string-artifactname
func (enclaveCtx *EnclaveContext) UploadFiles(pathToUpload string, artifactName string) (services.FilesArtifactUUID, services.FileArtifactName, error) {
	content, contentSize, _, err := utils.CompressPath(pathToUpload, enforceMaxFileSizeLimit)
	if err != nil {
		return "", "", stacktrace.Propagate(err,
			"There was an error compressing the file '%v' before upload",
			pathToUpload)
	}
	defer content.Close()

	// TODO: add content hash to API as well since we have it, as an optional field
	client, err := enclaveCtx.client.UploadFilesArtifact(context.Background())
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error was encountered initiating the data upload to the API Container.")
	}
	clientStream := grpc_file_streaming.NewClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse](client)
	response, err := clientStream.SendData(
		artifactName,
		content,
		contentSize,
		func(previousChunkHash string, contentChunk []byte) (*kurtosis_core_rpc_api_bindings.StreamedDataChunk, error) {
			return &kurtosis_core_rpc_api_bindings.StreamedDataChunk{
				Data:              contentChunk,
				PreviousChunkHash: previousChunkHash,
				Metadata: &kurtosis_core_rpc_api_bindings.DataChunkMetadata{
					Name: artifactName,
				},
			}, nil
		},
	)
	if err != nil {
		return "", "", stacktrace.Propagate(err, "An error was encountered while uploading data to the API Container.")
	}
	return services.FilesArtifactUUID(response.GetUuid()), services.FileArtifactName(response.GetName()), nil
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

	client, err := enclaveCtx.client.DownloadFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred initiating the download of files artifact '%v'", artifactIdentifier)
	}
	clientStream := grpc_file_streaming.NewClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, []byte](client)
	fileContent, err := clientStream.ReceiveData(
		artifactIdentifier,
		func(dataChunk *kurtosis_core_rpc_api_bindings.StreamedDataChunk) ([]byte, string, error) {
			return dataChunk.Data, dataChunk.PreviousChunkHash, nil
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred downloading files artifact '%v'", artifactIdentifier)
	}
	return fileContent, nil
}

func (enclaveCtx *EnclaveContext) InspectFilesArtifact(ctx context.Context, artifactName services.FileArtifactName) (*kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse, error) {
	// TODO(vcolombo): Add a more intuitive return type to this call instead of returning the RPC response
	response, err := enclaveCtx.client.InspectFilesArtifactContents(ctx, &kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest{
		FileNamesAndUuid: &kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid{
			FileName: string(artifactName),
			FileUuid: "",
		},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred inspecting file artifacts")
	}
	return response, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getexistingandhistoricalserviceidentifiers---serviceidentifiers-serviceidentifiers
func (enclaveCtx *EnclaveContext) GetExistingAndHistoricalServiceIdentifiers(ctx context.Context) (*services.ServiceIdentifiers, error) {
	response, err := enclaveCtx.client.GetExistingAndHistoricalServiceIdentifiers(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching existing and historical identifiers")
	}
	return services.NewServiceIdentifiers(enclaveCtx.enclaveName, response.AllIdentifiers), nil
}

// Docs available at https://docs.kurtosis.com/#getallfilesartifactnamesanduuids---filesartifactnameanduuid-filesartifactnamesanduuids
func (enclaveCtx *EnclaveContext) GetAllFilesArtifactNamesAndUuids(ctx context.Context) ([]*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid, error) {
	response, err := enclaveCtx.client.ListFilesArtifactNamesAndUuids(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching file names and uuids")
	}
	return response.GetFileNamesAndUuids(), nil
}

// Docs available at https://docs.kurtosis.com/sdk#connectservices
func (enclaveCtx *EnclaveContext) ConnectServices(ctx context.Context, connect kurtosis_core_rpc_api_bindings.Connect) error {
	args := binding_constructors.NewConnectServicesArgs(connect)
	_, err := enclaveCtx.client.ConnectServices(ctx, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error was encountered while sending the connect request to the API Container.")
	}
	return nil
}

// Docs available at https://docs.kurtosis.com/#getstarlarkrun
func (enclaveCtx *EnclaveContext) GetStarlarkRun(ctx context.Context) (*kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse, error) {
	response, err := enclaveCtx.client.GetStarlarkRun(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the last starlark run")
	}
	return response, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================

// convertApiPortsToServiceContextPorts returns a converted map where Port objects associated with strings in [apiPorts] are
// properly converted to PortSpec objects.
// Returns error if:
// - Any protocol associated with a port in [apiPorts] is invalid (eg. not currently supported).
// - Any port number associated with a port [apiPorts] is higher than the max port number.
func convertApiPortsToServiceContextPorts(apiPorts map[string]*kurtosis_core_rpc_api_bindings.Port) (map[string]*services.PortSpec, error) {
	result := map[string]*services.PortSpec{}
	for portId, apiPortSpec := range apiPorts {
		apiTransportProtocol := apiPortSpec.GetTransportProtocol()
		serviceCtxTransportProtocol := services.TransportProtocol(apiTransportProtocol)
		if !serviceCtxTransportProtocol.IsValid() {
			return nil, stacktrace.NewError("Received unrecognized protocol '%v' from the API", apiTransportProtocol)
		}
		portNumUint32 := apiPortSpec.GetNumber()
		if portNumUint32 > services.MaxPortNum {
			return nil, stacktrace.NewError(
				"Received port number '%v' from the API which is higher than the max allowed port number '%v'",
				portNumUint32,
				services.MaxPortNum,
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

func getErrFromStarlarkRunResult(result *StarlarkRunResult) error {
	if result.InterpretationError != nil {
		return stacktrace.NewError(result.InterpretationError.GetErrorMessage())
	}
	if len(result.ValidationErrors) > 0 {
		errorMessages := []string{}
		for _, validationErr := range result.ValidationErrors {
			errorMessages = append(errorMessages, validationErr.GetErrorMessage())
		}
		return stacktrace.NewError("Found %v validation errors: %v", len(result.ValidationErrors), strings.Join(errorMessages, "\n"))
	}
	if result.ExecutionError != nil {
		return stacktrace.NewError(result.ExecutionError.GetErrorMessage())
	}
	return nil
}

func (enclaveCtx *EnclaveContext) assembleRunStartosisPackageArg(
	packageName string,
	relativePathToMainFile string,
	mainFunctionName string,
	serializedParams string,
	dryRun bool,
	parallelism int32,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	cloudInstanceId string,
	cloudUserId string,
	imageDownloadMode kurtosis_core_rpc_api_bindings.ImageDownloadMode,
) (*kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, error) {

	return binding_constructors.NewRunStarlarkPackageArgs(packageName, relativePathToMainFile, mainFunctionName, serializedParams, dryRun, parallelism, experimentalFeatures, cloudInstanceId, cloudUserId, imageDownloadMode), nil
}

func (enclaveCtx *EnclaveContext) uploadStarlarkPackage(packageId string, packageRootPath string) error {
	logrus.Infof("Compressing package '%v' at '%v' for upload", packageId, packageRootPath)
	compressedModule, commpressedModuleSize, _, err := utils.CompressPath(packageRootPath, enforceMaxFileSizeLimit)
	if err != nil {
		return stacktrace.Propagate(err, "There was an error compressing module '%v' before upload", packageRootPath)
	}
	defer compressedModule.Close()
	logrus.Infof("Uploading and executing package '%v'", packageId)

	// TODO: add content hash to API here as well
	client, err := enclaveCtx.client.UploadStarlarkPackage(context.Background())
	if err != nil {
		return stacktrace.Propagate(err, "An error was encountered initiating the data upload to the API Container.")
	}
	clientStream := grpc_file_streaming.NewClientStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty](client)
	_, err = clientStream.SendData(
		packageId,
		compressedModule,
		commpressedModuleSize,
		func(previousChunkHash string, contentChunk []byte) (*kurtosis_core_rpc_api_bindings.StreamedDataChunk, error) {
			return &kurtosis_core_rpc_api_bindings.StreamedDataChunk{
				Data:              contentChunk,
				PreviousChunkHash: previousChunkHash,
				Metadata: &kurtosis_core_rpc_api_bindings.DataChunkMetadata{
					Name: packageId,
				},
			}, nil
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error was encountered while uploading data to the API Container.")
	}
	return nil
}

func (enclaveCtx *EnclaveContext) convertServiceInfoToServiceContext(serviceInfo *kurtosis_core_rpc_api_bindings.ServiceInfo) (*services.ServiceContext, error) {
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
		services.ServiceName(serviceInfo.Name),
		services.ServiceUUID(serviceInfo.ServiceUuid),
		serviceInfo.GetPrivateIpAddr(),
		serviceCtxPrivatePorts,
		serviceInfo.GetMaybePublicIpAddr(),
		serviceCtxPublicPorts,
	)

	return serviceContext, nil
}

func getKurtosisYaml(packageRootPath string) (*KurtosisYaml, error) {
	kurtosisYamlFilepath := path.Join(packageRootPath, kurtosisYamlFilename)

	kurtosisYaml, err := ParseKurtosisYaml(kurtosisYamlFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error parsing the '%v' at '%v'", kurtosisYamlFilename, packageRootPath)
	}
	return kurtosisYaml, nil
}

func isLocalDependencyReplace(replace string) bool {
	if strings.HasPrefix(replace, osPathSeparatorString) || strings.HasPrefix(replace, dotRelativePathIndicatorString) {
		return true
	}
	return false
}

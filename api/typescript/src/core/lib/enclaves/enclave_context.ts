/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

import {ok, err, Result, Err} from "neverthrow";
import * as jspb from "google-protobuf";
import type {
    Port,
    GetServicesArgs,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import { GrpcNodeApiContainerClient } from "./grpc_node_api_container_client";
import type { GenericApiContainerClient } from "./generic_api_container_client";
import {
    newDownloadFilesArtifactArgs,
    newGetServicesArgs,
    newStoreWebFilesArtifactArgs,
} from "../constructor_calls";
import type { FilesArtifactUUID } from "./files_artifact";
import type { ServiceName, ServiceUUID } from "../services/service";
import { ServiceContext } from "../services/service_context";
import { TransportProtocol, PortSpec, IsValidTransportProtocol, MAX_PORT_NUM } from "../services/port_spec";
import type { GenericPathJoiner } from "./generic_path_joiner";
import {GenericTgzArchiver} from "./generic_tgz_archiver";
import {
    ServiceInfo,
    RunStarlarkScriptArgs,
    RunStarlarkPackageArgs,
    FilesArtifactNameAndUuid,
    KurtosisFeatureFlag,
    ConnectServicesArgs,
    ConnectServicesResponse,
    Connect,
    GetStarlarkRunResponse,
} from "../../kurtosis_core_rpc_api_bindings/api_container_service_pb";
import * as path from "path";
import * as fs from 'fs';
import {parseKurtosisYaml, KurtosisYaml} from "./kurtosis_yaml";
import {Readable} from "stream";
import {readStreamContentUntilClosed, StarlarkRunResult} from "./starlark_run_blocking";
import {ServiceIdentifiers} from "../services/service_identifiers";
import {StarlarkRunConfig} from "./starlark_run_config"

export type EnclaveUUID = string;

export const KURTOSIS_YAML_FILENAME = "kurtosis.yml";

const OS_PATH_SEPARATOR_STRING = "/"

const DOT_RELATIVE_PATH_INDICATOR_STRING = "."

// required to get around the "only Github URLs" validation
const composePackageIdPlaceholder = 'github.com/NOTIONAL_USER/COMPOSE-PACKAGE'


// TODO Remove this once package ID is detected ONLY the APIC side (i.e. the CLI doesn't need to tell the APIC what package ID it's using)
// Doing so requires that we upload completely anonymous packages to the APIC, and it figures things out from there
let supportedDockerComposeYmlFilenames = [
    "compose.yml",
    "compose.yaml",
    "docker-compose.yml",
    "docker-compose.yaml",
    "docker_compose.yml",
    "docker_compose.yaml",
]

// Docs available at https://docs.kurtosis.com/sdk/#enclavecontext
export class EnclaveContext {

    private readonly backend: GenericApiContainerClient
    private readonly pathJoiner: GenericPathJoiner
    private readonly genericTgzArchiver: GenericTgzArchiver

    private constructor(backend: GenericApiContainerClient, pathJoiner: GenericPathJoiner,
                        genericTgzArchiver: GenericTgzArchiver){
        this.backend = backend;
        this.pathJoiner = pathJoiner;
        this.genericTgzArchiver = genericTgzArchiver
    }

    public static async newGrpcNodeEnclaveContext(
        ipAddress: string,
        apiContainerGrpcPortNum: number,
        enclaveUuid: string,
        enclaveName: string,
    ): Promise<Result<EnclaveContext, Error>> {

        let genericApiContainerClient: GenericApiContainerClient
        let genericTgzArchiver: GenericTgzArchiver
        let pathJoiner: GenericPathJoiner
        //TODO Pull things that can't throw an error out of try statement.
        try {
            pathJoiner = await import( /* webpackIgnore: true */ "path")
            const grpc_node = await import( /* webpackIgnore: true */ "@grpc/grpc-js")
            const apiContainerServiceNode = await import( /* webpackIgnore: true */ "../../kurtosis_core_rpc_api_bindings/api_container_service_grpc_pb")

            const apiContainerGrpcUrl: string = `${ipAddress}:${apiContainerGrpcPortNum}`
            const apiContainerClient = new apiContainerServiceNode.ApiContainerServiceClient(apiContainerGrpcUrl, grpc_node.credentials.createInsecure());
            genericApiContainerClient = new GrpcNodeApiContainerClient(apiContainerClient, enclaveUuid, enclaveName)

            const nodeTgzArchiver = await import(/* webpackIgnore: true */ "./node_tgz_archiver")
            genericTgzArchiver = new nodeTgzArchiver.NodeTgzArchiver()
        }catch(error) {
            if (error instanceof Error) {
                return err(error);
            }
            return err(new Error(
                "An unknown exception value was thrown during creation of the API container client that wasn't an error: " + error
            ));
        }

        const enclaveContext = new EnclaveContext(genericApiContainerClient, pathJoiner, genericTgzArchiver);
        return ok(enclaveContext)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getenclaveuuid---enclaveuuid
    public getEnclaveUuid(): EnclaveUUID {
        return this.backend.getEnclaveUuid();
    }

    // Docs available at https://docs.kurtosis.com/sdk/#getenclavename---string
    public getEnclaveName(): string {
        return this.backend.getEnclaveName();
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptstring-serializedstarlarkscript-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkScript(
        serializedStartosisScript: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<Readable, Error>> {
        const args = new RunStarlarkScriptArgs();
        args.setSerializedScript(serializedStartosisScript)
        args.setSerializedParams(runConfig.serializedParams)
        args.setDryRun(runConfig.dryRun)
        args.setMainFunctionName(runConfig.mainFunctionName)
        args.setExperimentalFeaturesList(runConfig.experimentalFeatureFlags)
        args.setCloudInstanceId(runConfig.cloudInstanceId)
        args.setCloudUserId(runConfig.cloudUserId)
        const scriptRunResult : Result<Readable, Error> = await this.backend.runStarlarkScript(args)
        if (scriptRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark script \n${scriptRunResult.error}`))
        }
        return ok(scriptRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkscriptblockingstring-serializedstarlarkscript-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkScriptBlocking(
        serializedStartosisScript: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkScript(serializedStartosisScript, runConfig)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackagestring-packagerootpath-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkPackage(
        packageRootPath: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<Readable, Error>> {
        const packageNameAndReplaceOptionsResult = await this.getPackageNameAndReplaceOptions(packageRootPath);
        if (packageNameAndReplaceOptionsResult.isErr()) {
            return err(new Error(`Unexpected error occurred while trying to get package name and replace options:\n${packageNameAndReplaceOptionsResult.error}`))
        }
        const [packageName, packageReplaceOptions] = packageNameAndReplaceOptionsResult.value;

        const args = await this.assembleRunStarlarkPackageArg(
            packageName,
            runConfig.relativePathToMainFile,
            runConfig.mainFunctionName,
            runConfig.serializedParams,
            runConfig.dryRun,
            runConfig.cloudInstanceId,
            runConfig.cloudUserId)
        if (args.isErr()) {
            return err(new Error(`Unexpected error while assembling arguments to pass to the Starlark executor \n${args.error}`))
        }

        const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(packageRootPath)
        if (archiverResponse.isErr()){
            return err(new Error(`Unexpected error while creating the package's tgs file from '${packageRootPath}'\n${archiverResponse.error}`))
        }

        const uploadStarlarkPackageResponse = await this.backend.uploadStarlarkPackage(packageName, archiverResponse.value)
        if (uploadStarlarkPackageResponse.isErr()){
            return err(new Error(`Unexpected error while uploading Starlark package '${packageName}'\n${uploadStarlarkPackageResponse.error}`))
        }

        if (packageReplaceOptions !== undefined && packageReplaceOptions.size > 0) {
            const uploadLocalStarlarkPackageDependenciesResponse = await this.uploadLocalStarlarkPackageDependencies(packageRootPath, packageReplaceOptions)
            if (uploadLocalStarlarkPackageDependenciesResponse.isErr()) {
                return err(new Error(`Unexpected error while uploading local Starlark package dependencies '${packageReplaceOptions}' from '${packageRootPath}' \n${uploadLocalStarlarkPackageDependenciesResponse.error}`))
            }
        }

        const packageRunResult : Result<Readable, Error> = await this.backend.runStarlarkPackage(args.value)
        if (packageRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark package \n${packageRunResult.error}`))
        }
        return ok(packageRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkpackageblockingstring-packagerootpath-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkPackageBlocking(
        packageRootPath: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkPackage(packageRootPath, runConfig)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runremotestarlarkpackagestring-packageid-string-serializedparams-boolean-dryrun---streamstarlarkrunresponseline-responselines-error-error
    public async runStarlarkRemotePackage(
        packageId: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<Readable, Error>> {
        const args = new RunStarlarkPackageArgs();
        args.setPackageId(packageId)
        args.setDryRun(runConfig.dryRun)
        args.setSerializedParams(runConfig.serializedParams)
        args.setRemote(true)
        args.setRelativePathToMainFile(runConfig.relativePathToMainFile)
        args.setMainFunctionName(runConfig.mainFunctionName)
        args.setCloudInstanceId(runConfig.cloudInstanceId)
        args.setCloudUserId(runConfig.cloudUserId)
        const remotePackageRunResult : Result<Readable, Error> = await this.backend.runStarlarkPackage(args)
        if (remotePackageRunResult.isErr()) {
            return err(new Error(`Unexpected error happened executing Starlark package \n${remotePackageRunResult.error}`))
        }
        return ok(remotePackageRunResult.value)
    }

    // Docs available at https://docs.kurtosis.com/sdk/#runstarlarkremotepackageblockingstring-packageid-string-serializedparams-boolean-dryrun---starlarkrunresult-runresult-error-error
    public async runStarlarkRemotePackageBlocking(
        packageId: string,
        runConfig: StarlarkRunConfig,
    ): Promise<Result<StarlarkRunResult, Error>> {
        const runAsyncResponse = await this.runStarlarkRemotePackage(packageId, runConfig)
        if (runAsyncResponse.isErr()) {
            return err(runAsyncResponse.error)
        }
        const fullRunResult = await readStreamContentUntilClosed(runAsyncResponse.value)
        return ok(fullRunResult)
    }

    // Docs available at https://docs.kurtosis.com/sdk#getservicecontextstring-serviceidentifier---servicecontext-servicecontext
    public async getServiceContext(serviceIdentifier: string): Promise<Result<ServiceContext, Error>> {
        const serviceArgMap = new Map<string, boolean>()
        serviceArgMap.set(serviceIdentifier, true)
        const getServiceInfoArgs: GetServicesArgs = newGetServicesArgs(serviceArgMap);

        const getServicesResult = await this.backend.getServices(getServiceInfoArgs)
        if(getServicesResult.isErr()){
            return err(getServicesResult.error)
        }

        const serviceInfo = getServicesResult.value.getServiceInfoMap().get(serviceIdentifier)
        if(!serviceInfo) {
            return err(new Error(
                    "Failed to retrieve service information for service " + serviceIdentifier
            ))
        }
        if (serviceInfo.getPrivateIpAddr() === "") {
            return err(new Error(
                    "Kurtosis API reported an empty private IP address for service " + serviceIdentifier +  " - this should never happen, and is a bug with Kurtosis!",
                )
            );
        }

        const resultConvertServiceCtxPrivatePorts: Result<Map<string, PortSpec>,Error> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getPrivatePortsMap(),
        );
        if (resultConvertServiceCtxPrivatePorts.isErr()){
            return err(resultConvertServiceCtxPrivatePorts.error);
        }
        const serviceCtxPrivatePorts: Map<string, PortSpec> = resultConvertServiceCtxPrivatePorts.value;
        const resultConvertServiceCtxPublicPorts: Result<Map<string, PortSpec>,Error> = EnclaveContext.convertApiPortsToServiceContextPorts(
            serviceInfo.getMaybePublicPortsMap(),
        );
        if (resultConvertServiceCtxPublicPorts.isErr()){
            return err(resultConvertServiceCtxPublicPorts.error);
        }
        const serviceCtxPublicPorts: Map<string, PortSpec> = resultConvertServiceCtxPublicPorts.value;

        const serviceContext: ServiceContext = new ServiceContext(
            this.backend,
            serviceIdentifier,
            serviceInfo.getServiceUuid(),
            serviceInfo.getPrivateIpAddr(),
            serviceCtxPrivatePorts,
            serviceInfo.getMaybePublicIpAddr(),
            serviceCtxPublicPorts,
            serviceInfo.getServiceStatus(),
            serviceInfo.getContainer(),
        );

        return ok(serviceContext);
    }

    // TODO: Add getServiceContexts

    // Docs available at https://docs.kurtosis.com/sdk#getservices---mapservicename--serviceuuid-serviceidentifiers
    public async getServices(): Promise<Result<Map<ServiceName, ServiceUUID>, Error>> {
        const getAllServicesArgMap: Map<string, boolean> = new Map<string,boolean>()
        const emptyGetServicesArg: GetServicesArgs = newGetServicesArgs(getAllServicesArgMap)

        const getServicesResponseResult = await this.backend.getServices(emptyGetServicesArg)
        if(getServicesResponseResult.isErr()){
            return err(getServicesResponseResult.error)
        }

        const getServicesResponse = getServicesResponseResult.value

        const serviceInfos: Map<ServiceName, ServiceUUID> = new Map<ServiceName, ServiceUUID>()
        getServicesResponse.getServiceInfoMap().forEach((value: ServiceInfo, key: string) => {
            serviceInfos.set(key, value.getServiceUuid())
        });
        return ok(serviceInfos)
    }

    // Docs available at https://docs.kurtosis.com/sdk#uploadfilesstring-pathtoupload-string-artifactname
    public async uploadFiles(pathToArchive: string, name: string): Promise<Result<FilesArtifactUUID, Error>>  {
        const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(pathToArchive)
        if (archiverResponse.isErr()){
            return err(archiverResponse.error)
        }

        const uploadResult = await this.backend.uploadFiles(name, archiverResponse.value)
        if (uploadResult.isErr()){
            return err(uploadResult.error)
        }

        return ok(uploadResult.value.getUuid())
    }

    // Docs available at https://docs.kurtosis.com/sdk#storewebfilesstring-urltodownload-string-artifactname
    public async storeWebFiles(url: string, name: string): Promise<Result<FilesArtifactUUID, Error>> {
        const args = newStoreWebFilesArtifactArgs(url, name);
        const storeWebFilesArtifactResponseResult = await this.backend.storeWebFilesArtifact(args)
        if (storeWebFilesArtifactResponseResult.isErr()) {
            return err(storeWebFilesArtifactResponseResult.error)
        }
        const storeWebFilesArtifactResponse = storeWebFilesArtifactResponseResult.value;
        return ok(storeWebFilesArtifactResponse.getUuid())
    }

    // Docs available at https://docs.kurtosis.com/sdk#downloadfilesartifact-fileidentifier-string
    public async downloadFilesArtifact(identifier: string): Promise<Result<Uint8Array, Error>> {
        const args = newDownloadFilesArtifactArgs(identifier);
        const downloadFilesArtifactResponseResult = await this.backend.downloadFilesArtifact(args)
        if (downloadFilesArtifactResponseResult.isErr()) {
            return err(downloadFilesArtifactResponseResult.error)
        }
        const downloadFilesArtifactResponse = downloadFilesArtifactResponseResult.value;
        return ok(downloadFilesArtifactResponse)
    }

    // Docs available at https://docs.kurtosis.com/sdk#getexistingandhistoricalserviceidentifiers---serviceidentifiers-serviceidentifiers
    public async getExistingAndHistoricalServiceIdentifiers(): Promise<Result<ServiceIdentifiers, Error>> {
        const getExistingAndHistoricalServiceIdentifiersResponseResult = await this.backend.getExistingAndHistoricalServiceIdentifiers()
        if (getExistingAndHistoricalServiceIdentifiersResponseResult.isErr()) {
            return err(getExistingAndHistoricalServiceIdentifiersResponseResult.error);
        }

        const getExistingAndHistoricalIdentifiersValue = getExistingAndHistoricalServiceIdentifiersResponseResult.value
        return ok(new ServiceIdentifiers(getExistingAndHistoricalIdentifiersValue.getAllidentifiersList()))
    }

    // Docs available at https://docs.kurtosis.com/#getallfilesartifactnamesanduuids---filesartifactnameanduuid-filesartifactnamesanduuids
    public async getAllFilesArtifactNamesAndUuids(): Promise<Result<FilesArtifactNameAndUuid[], Error>> {
        const getAllFilesArtifactsNamesAndUuidsResponseResult = await this.backend.getAllFilesArtifactNamesAndUuids()
        if (getAllFilesArtifactsNamesAndUuidsResponseResult.isErr()) {
            return err(getAllFilesArtifactsNamesAndUuidsResponseResult.error)
        }

        const getAllFilesArtifactsNamesAndUuidsResponseValue = getAllFilesArtifactsNamesAndUuidsResponseResult.value
        return ok(getAllFilesArtifactsNamesAndUuidsResponseValue.getFileNamesAndUuidsList())
    }

    // Docs available at https://docs.kurtosis.com/sdk#connectservices-connect-string
    public async connectServices(connect: Connect): Promise<Result<ConnectServicesResponse, Error>> {
        const args = new ConnectServicesArgs()
        args.setConnect(connect)
        const responseResult = await this.backend.connectServices(args)
        if (responseResult.isErr()) {
            return err(responseResult.error)
        }
        const response = responseResult.value;
        return ok(response)
    }

    // Docs available at https://docs.kurtosis.com/sdk#getstarlarkrun
    public async getStarlarkRun(): Promise<Result<GetStarlarkRunResponse, Error>> {
        const responseResult = await this.backend.getStarlarkRun()
        if (responseResult.isErr()) {
            return err(responseResult.error)
        }
        const response = responseResult.value;
        return ok(response)
    }

    // ====================================================================================================
    //                                       Private helper functions
    // ====================================================================================================
    // Determines the package name and replace options based on [packageRootPath]
    // If a kurtosis.yml is detected, package is a kurtosis package
    // If a valid [supportedDockerComposeYaml] is detected, package is a docker compose package
    private async getPackageNameAndReplaceOptions(packageRootPath: string): Promise<Result<[string, Map<string, string>], Error>>
    {
        let packageName: string;
        let packageReplaceOptions: Map<string, string>;

        // Use kurtosis package if it exists
        if (fs.existsSync(path.join(packageRootPath, KURTOSIS_YAML_FILENAME))) {
            const kurtosisYmlResult = await this.getKurtosisYaml(packageRootPath)
            if (kurtosisYmlResult.isErr()) {
                return err(new Error(`Unexpected error while getting the Kurtosis yaml file from path '${packageRootPath}'`))
            }
            const kurtosisYml: KurtosisYaml = kurtosisYmlResult.value
            packageName = kurtosisYml.name;
            packageReplaceOptions = kurtosisYml.packageReplaceOptions;
        } else {
            // Use compose package if it exists
            let composeAbsFilepath = '';
            for (const candidateComposeFilename of supportedDockerComposeYmlFilenames) {
                const candidateComposeAbsFilepath = path.join(packageRootPath, candidateComposeFilename);
                if (fs.existsSync(candidateComposeAbsFilepath)) {
                    composeAbsFilepath = candidateComposeAbsFilepath;
                    break;
                }
            }
            if (composeAbsFilepath === '') {
                return err(new Error(
                    `Neither a '${KURTOSIS_YAML_FILENAME}' file nor one of the default Compose files (${supportedDockerComposeYmlFilenames.join(', ')}) was found in the package root; at least one of these is required`,
                ));
            }
            packageName = composePackageIdPlaceholder;
            packageReplaceOptions = new Map<string, string>();
        }

        return ok([packageName, packageReplaceOptions]);
    }

    // convertApiPortsToServiceContextPorts returns a converted map where Port objects associated with strings in [apiPorts] are
    // properly converted to PortSpec objects.
    // Returns error if:
    // - Any protocol associated with a port in [apiPorts] is invalid (eg. not currently supported).
    // - Any port number associated with a port [apiPorts] is higher than the max port number.
    private static convertApiPortsToServiceContextPorts(apiPorts: jspb.Map<string, Port>): Result<Map<string, PortSpec>,Error> {
        const result: Map<string, PortSpec> = new Map();
        for (const [portId, apiPortSpec] of apiPorts.entries()) {
            const portProtocol: TransportProtocol = apiPortSpec.getTransportProtocol();
            if (!IsValidTransportProtocol(portProtocol)){
                return err(new Error("Received unrecognized protocol '"+ portProtocol + "' from the API"))
            }
            const portNum: number = apiPortSpec.getNumber();
            if (portNum > MAX_PORT_NUM){
                return err(new Error("Received port number '"+ portNum +"' from the API which is higher than the max allowed port number + '"+ MAX_PORT_NUM + "'"))
            }
            const portSpec = new PortSpec(portNum, portProtocol, apiPortSpec.getMaybeApplicationProtocol());
            result.set(portId, portSpec)
        }
        return ok(result);
    }

    private async assembleRunStarlarkPackageArg(
        packageName: string,
        relativePathToMainFile: string,
        mainFunctionName: string,
        serializedParams: string,
        dryRun: boolean,
        cloudInstanceId: string,
        cloudUserId: string,
        ): Promise<Result<RunStarlarkPackageArgs, Error>> {

        const args = new RunStarlarkPackageArgs;
        args.setPackageId(packageName)
        args.setSerializedParams(serializedParams)
        args.setDryRun(dryRun)
        args.setRelativePathToMainFile(relativePathToMainFile)
        args.setMainFunctionName(mainFunctionName)
        args.setCloudInstanceId(cloudInstanceId)
        args.setCloudUserId(cloudUserId)
        return ok(args)
    }

    private async getKurtosisYaml(packageRootPath: string): Promise<Result<KurtosisYaml, Error>> {
        const kurtosisYamlFilepath = path.join(packageRootPath, KURTOSIS_YAML_FILENAME)

        const resultParseKurtosisYaml = await parseKurtosisYaml(kurtosisYamlFilepath)
        if (resultParseKurtosisYaml.isErr()) {
            return err(resultParseKurtosisYaml.error)
        }
        const kurtosisYaml = resultParseKurtosisYaml.value

        return ok(kurtosisYaml)
    }


    private async uploadLocalStarlarkPackageDependencies(
        packageRootPath: string,
        packageReplaceOptions: Map<string, string>,
    ): Promise<Result<null, Error>> {
        for (const [dependencyPackageId, replaceOption] of packageReplaceOptions.entries()) {
            if (this.isLocalDependencyReplace(replaceOption)) {
                const localPackagePath: string = path.join(packageRootPath, replaceOption)

                const archiverResponse = await this.genericTgzArchiver.createTgzByteArray(localPackagePath)
                if (archiverResponse.isErr()){
                    return err(archiverResponse.error)
                }

                const uploadStarlarkPackageResponse = await this.backend.uploadStarlarkPackage(dependencyPackageId, archiverResponse.value)
                if (uploadStarlarkPackageResponse.isErr()){
                    return err(uploadStarlarkPackageResponse.error)
                }
                return ok(null)
            }
        }
        return ok(null)
    }

    private isLocalDependencyReplace(replace: string): boolean {
        if (replace.startsWith(OS_PATH_SEPARATOR_STRING) || replace.startsWith(DOT_RELATIVE_PATH_INDICATOR_STRING)) {
            return true
        }
        return false
    }
}

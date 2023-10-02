// @generated by protoc-gen-connect-es v0.12.0 with parameter "target=js+dts"
// @generated from file api_container_service.proto (package api_container_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { ConnectServicesArgs, ConnectServicesResponse, DownloadFilesArtifactArgs, ExecCommandArgs, ExecCommandResponse, GetExistingAndHistoricalServiceIdentifiersResponse, GetServicesArgs, GetServicesResponse, GetStarlarkRunResponse, InspectFilesArtifactContentsRequest, InspectFilesArtifactContentsResponse, ListFilesArtifactNamesAndUuidsResponse, RunStarlarkPackageArgs, RunStarlarkScriptArgs, StarlarkRunResponseLine, StoreFilesArtifactFromServiceArgs, StoreFilesArtifactFromServiceResponse, StoreWebFilesArtifactArgs, StoreWebFilesArtifactResponse, StreamedDataChunk, UploadFilesArtifactResponse, WaitForHttpGetEndpointAvailabilityArgs, WaitForHttpPostEndpointAvailabilityArgs } from "./api_container_service_pb.js";
import { Empty, MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service api_container_api.ApiContainerService
 */
export declare const ApiContainerService: {
  readonly typeName: "api_container_api.ApiContainerService",
  readonly methods: {
    /**
     * Executes a Starlark script on the user's behalf
     *
     * @generated from rpc api_container_api.ApiContainerService.RunStarlarkScript
     */
    readonly runStarlarkScript: {
      readonly name: "RunStarlarkScript",
      readonly I: typeof RunStarlarkScriptArgs,
      readonly O: typeof StarlarkRunResponseLine,
      readonly kind: MethodKind.ServerStreaming,
    },
    /**
     * Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
     *
     * @generated from rpc api_container_api.ApiContainerService.UploadStarlarkPackage
     */
    readonly uploadStarlarkPackage: {
      readonly name: "UploadStarlarkPackage",
      readonly I: typeof StreamedDataChunk,
      readonly O: typeof Empty,
      readonly kind: MethodKind.ClientStreaming,
    },
    /**
     * Executes a Starlark script on the user's behalf
     *
     * @generated from rpc api_container_api.ApiContainerService.RunStarlarkPackage
     */
    readonly runStarlarkPackage: {
      readonly name: "RunStarlarkPackage",
      readonly I: typeof RunStarlarkPackageArgs,
      readonly O: typeof StarlarkRunResponseLine,
      readonly kind: MethodKind.ServerStreaming,
    },
    /**
     * Returns the IDs of the current services in the enclave
     *
     * @generated from rpc api_container_api.ApiContainerService.GetServices
     */
    readonly getServices: {
      readonly name: "GetServices",
      readonly I: typeof GetServicesArgs,
      readonly O: typeof GetServicesResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Returns information about all existing & historical services
     *
     * @generated from rpc api_container_api.ApiContainerService.GetExistingAndHistoricalServiceIdentifiers
     */
    readonly getExistingAndHistoricalServiceIdentifiers: {
      readonly name: "GetExistingAndHistoricalServiceIdentifiers",
      readonly I: typeof Empty,
      readonly O: typeof GetExistingAndHistoricalServiceIdentifiersResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Executes the given command inside a running container
     *
     * @generated from rpc api_container_api.ApiContainerService.ExecCommand
     */
    readonly execCommand: {
      readonly name: "ExecCommand",
      readonly I: typeof ExecCommandArgs,
      readonly O: typeof ExecCommandResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
     *
     * @generated from rpc api_container_api.ApiContainerService.WaitForHttpGetEndpointAvailability
     */
    readonly waitForHttpGetEndpointAvailability: {
      readonly name: "WaitForHttpGetEndpointAvailability",
      readonly I: typeof WaitForHttpGetEndpointAvailabilityArgs,
      readonly O: typeof Empty,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
     *
     * @generated from rpc api_container_api.ApiContainerService.WaitForHttpPostEndpointAvailability
     */
    readonly waitForHttpPostEndpointAvailability: {
      readonly name: "WaitForHttpPostEndpointAvailability",
      readonly I: typeof WaitForHttpPostEndpointAvailabilityArgs,
      readonly O: typeof Empty,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Uploads a files artifact to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.UploadFilesArtifact
     */
    readonly uploadFilesArtifact: {
      readonly name: "UploadFilesArtifact",
      readonly I: typeof StreamedDataChunk,
      readonly O: typeof UploadFilesArtifactResponse,
      readonly kind: MethodKind.ClientStreaming,
    },
    /**
     * Downloads a files artifact from the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.DownloadFilesArtifact
     */
    readonly downloadFilesArtifact: {
      readonly name: "DownloadFilesArtifact",
      readonly I: typeof DownloadFilesArtifactArgs,
      readonly O: typeof StreamedDataChunk,
      readonly kind: MethodKind.ServerStreaming,
    },
    /**
     * Tells the API container to download a files artifact from the web to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.StoreWebFilesArtifact
     */
    readonly storeWebFilesArtifact: {
      readonly name: "StoreWebFilesArtifact",
      readonly I: typeof StoreWebFilesArtifactArgs,
      readonly O: typeof StoreWebFilesArtifactResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Tells the API container to copy a files artifact from a service to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.StoreFilesArtifactFromService
     */
    readonly storeFilesArtifactFromService: {
      readonly name: "StoreFilesArtifactFromService",
      readonly I: typeof StoreFilesArtifactFromServiceArgs,
      readonly O: typeof StoreFilesArtifactFromServiceResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc api_container_api.ApiContainerService.ListFilesArtifactNamesAndUuids
     */
    readonly listFilesArtifactNamesAndUuids: {
      readonly name: "ListFilesArtifactNamesAndUuids",
      readonly I: typeof Empty,
      readonly O: typeof ListFilesArtifactNamesAndUuidsResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc api_container_api.ApiContainerService.InspectFilesArtifactContents
     */
    readonly inspectFilesArtifactContents: {
      readonly name: "InspectFilesArtifactContents",
      readonly I: typeof InspectFilesArtifactContentsRequest,
      readonly O: typeof InspectFilesArtifactContentsResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * User services port forwarding
     *
     * @generated from rpc api_container_api.ApiContainerService.ConnectServices
     */
    readonly connectServices: {
      readonly name: "ConnectServices",
      readonly I: typeof ConnectServicesArgs,
      readonly O: typeof ConnectServicesResponse,
      readonly kind: MethodKind.Unary,
    },
    /**
     * Get last Starlark run
     *
     * @generated from rpc api_container_api.ApiContainerService.GetStarlarkRun
     */
    readonly getStarlarkRun: {
      readonly name: "GetStarlarkRun",
      readonly I: typeof Empty,
      readonly O: typeof GetStarlarkRunResponse,
      readonly kind: MethodKind.Unary,
    },
  }
};


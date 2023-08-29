// @generated by protoc-gen-connect-es v0.12.0 with parameter "target=js+dts"
// @generated from file api_container_service.proto (package api_container_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { DownloadFilesArtifactArgs, ExecCommandArgs, ExecCommandResponse, GetExistingAndHistoricalServiceIdentifiersResponse, GetServicesArgs, GetServicesResponse, InspectFilesArtifactContentsRequest, InspectFilesArtifactContentsResponse, ListFilesArtifactNamesAndUuidsResponse, RunStarlarkPackageArgs, RunStarlarkScriptArgs, StarlarkRunResponseLine, StoreFilesArtifactFromServiceArgs, StoreFilesArtifactFromServiceResponse, StoreWebFilesArtifactArgs, StoreWebFilesArtifactResponse, StreamedDataChunk, UploadFilesArtifactResponse, WaitForHttpGetEndpointAvailabilityArgs, WaitForHttpPostEndpointAvailabilityArgs } from "./api_container_service_pb.js";
import { Empty, MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service api_container_api.ApiContainerService
 */
export const ApiContainerService = {
  typeName: "api_container_api.ApiContainerService",
  methods: {
    /**
     * Executes a Starlark script on the user's behalf
     *
     * @generated from rpc api_container_api.ApiContainerService.RunStarlarkScript
     */
    runStarlarkScript: {
      name: "RunStarlarkScript",
      I: RunStarlarkScriptArgs,
      O: StarlarkRunResponseLine,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Uploads a Starlark package. This step is required before the package can be executed with RunStarlarkPackage
     *
     * @generated from rpc api_container_api.ApiContainerService.UploadStarlarkPackage
     */
    uploadStarlarkPackage: {
      name: "UploadStarlarkPackage",
      I: StreamedDataChunk,
      O: Empty,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Executes a Starlark script on the user's behalf
     *
     * @generated from rpc api_container_api.ApiContainerService.RunStarlarkPackage
     */
    runStarlarkPackage: {
      name: "RunStarlarkPackage",
      I: RunStarlarkPackageArgs,
      O: StarlarkRunResponseLine,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Returns the IDs of the current services in the enclave
     *
     * @generated from rpc api_container_api.ApiContainerService.GetServices
     */
    getServices: {
      name: "GetServices",
      I: GetServicesArgs,
      O: GetServicesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Returns information about all existing & historical services
     *
     * @generated from rpc api_container_api.ApiContainerService.GetExistingAndHistoricalServiceIdentifiers
     */
    getExistingAndHistoricalServiceIdentifiers: {
      name: "GetExistingAndHistoricalServiceIdentifiers",
      I: Empty,
      O: GetExistingAndHistoricalServiceIdentifiersResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Executes the given command inside a running container
     *
     * @generated from rpc api_container_api.ApiContainerService.ExecCommand
     */
    execCommand: {
      name: "ExecCommand",
      I: ExecCommandArgs,
      O: ExecCommandResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Block until the given HTTP endpoint returns available, calling it through a HTTP Get request
     *
     * @generated from rpc api_container_api.ApiContainerService.WaitForHttpGetEndpointAvailability
     */
    waitForHttpGetEndpointAvailability: {
      name: "WaitForHttpGetEndpointAvailability",
      I: WaitForHttpGetEndpointAvailabilityArgs,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * Block until the given HTTP endpoint returns available, calling it through a HTTP Post request
     *
     * @generated from rpc api_container_api.ApiContainerService.WaitForHttpPostEndpointAvailability
     */
    waitForHttpPostEndpointAvailability: {
      name: "WaitForHttpPostEndpointAvailability",
      I: WaitForHttpPostEndpointAvailabilityArgs,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * Uploads a files artifact to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.UploadFilesArtifact
     */
    uploadFilesArtifact: {
      name: "UploadFilesArtifact",
      I: StreamedDataChunk,
      O: UploadFilesArtifactResponse,
      kind: MethodKind.ClientStreaming,
    },
    /**
     * Downloads a files artifact from the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.DownloadFilesArtifact
     */
    downloadFilesArtifact: {
      name: "DownloadFilesArtifact",
      I: DownloadFilesArtifactArgs,
      O: StreamedDataChunk,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * Tells the API container to download a files artifact from the web to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.StoreWebFilesArtifact
     */
    storeWebFilesArtifact: {
      name: "StoreWebFilesArtifact",
      I: StoreWebFilesArtifactArgs,
      O: StoreWebFilesArtifactResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Tells the API container to copy a files artifact from a service to the Kurtosis File System
     *
     * @generated from rpc api_container_api.ApiContainerService.StoreFilesArtifactFromService
     */
    storeFilesArtifactFromService: {
      name: "StoreFilesArtifactFromService",
      I: StoreFilesArtifactFromServiceArgs,
      O: StoreFilesArtifactFromServiceResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc api_container_api.ApiContainerService.ListFilesArtifactNamesAndUuids
     */
    listFilesArtifactNamesAndUuids: {
      name: "ListFilesArtifactNamesAndUuids",
      I: Empty,
      O: ListFilesArtifactNamesAndUuidsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc api_container_api.ApiContainerService.InspectFilesArtifactContents
     */
    inspectFilesArtifactContents: {
      name: "InspectFilesArtifactContents",
      I: InspectFilesArtifactContentsRequest,
      O: InspectFilesArtifactContentsResponse,
      kind: MethodKind.Unary,
    },
  }
};


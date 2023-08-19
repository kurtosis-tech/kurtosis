// @generated by protoc-gen-connect-es v0.12.0 with parameter "target=ts"
// @generated from file kurtosis_enclave_manager_api.proto (package kurtosis_enclave_manager, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { GetListFilesArtifactNamesAndUuidsRequest, GetServicesRequest, HealthCheckRequest, HealthCheckResponse, RunStarlarkPackageRequest } from "./kurtosis_enclave_manager_api_pb.js";
import { Empty, MethodKind } from "@bufbuild/protobuf";
import { CreateEnclaveArgs, CreateEnclaveResponse, GetEnclavesResponse, GetServiceLogsArgs, GetServiceLogsResponse } from "./engine_service_pb.js";
import { GetServicesResponse, ListFilesArtifactNamesAndUuidsResponse, StarlarkRunResponseLine } from "./api_container_service_pb.js";

/**
 * @generated from service kurtosis_enclave_manager.KurtosisEnclaveManagerServer
 */
export const KurtosisEnclaveManagerServer = {
  typeName: "kurtosis_enclave_manager.KurtosisEnclaveManagerServer",
  methods: {
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.Check
     */
    check: {
      name: "Check",
      I: HealthCheckRequest,
      O: HealthCheckResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetEnclaves
     */
    getEnclaves: {
      name: "GetEnclaves",
      I: Empty,
      O: GetEnclavesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServices
     */
    getServices: {
      name: "GetServices",
      I: GetServicesRequest,
      O: GetServicesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.GetServiceLogs
     */
    getServiceLogs: {
      name: "GetServiceLogs",
      I: GetServiceLogsArgs,
      O: GetServiceLogsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.ListFilesArtifactNamesAndUuids
     */
    listFilesArtifactNamesAndUuids: {
      name: "ListFilesArtifactNamesAndUuids",
      I: GetListFilesArtifactNamesAndUuidsRequest,
      O: ListFilesArtifactNamesAndUuidsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.RunStarlarkPackage
     */
    runStarlarkPackage: {
      name: "RunStarlarkPackage",
      I: RunStarlarkPackageRequest,
      O: StarlarkRunResponseLine,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc kurtosis_enclave_manager.KurtosisEnclaveManagerServer.CreateEnclave
     */
    createEnclave: {
      name: "CreateEnclave",
      I: CreateEnclaveArgs,
      O: CreateEnclaveResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;


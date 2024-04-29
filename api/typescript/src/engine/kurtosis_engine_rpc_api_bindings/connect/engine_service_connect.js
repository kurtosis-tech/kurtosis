// @generated by protoc-gen-connect-es v1.4.0 with parameter "target=js+dts"
// @generated from file engine_service.proto (package engine_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { Empty, MethodKind } from "@bufbuild/protobuf";
import { CleanArgs, CleanResponse, CreateEnclaveArgs, CreateEnclaveResponse, DestroyEnclaveArgs, GetEnclavesResponse, GetEngineInfoResponse, GetExistingAndHistoricalEnclaveIdentifiersResponse, GetServiceLogsArgs, GetServiceLogsResponse, StopEnclaveArgs } from "./engine_service_pb.js";

/**
 * @generated from service engine_api.EngineService
 */
export const EngineService = {
  typeName: "engine_api.EngineService",
  methods: {
    /**
     * Endpoint for getting information about the engine, which is also what we use to verify that the engine has become available
     *
     * @generated from rpc engine_api.EngineService.GetEngineInfo
     */
    getEngineInfo: {
      name: "GetEngineInfo",
      I: Empty,
      O: GetEngineInfoResponse,
      kind: MethodKind.Unary,
    },
    /**
     * ==============================================================================================
     *                                   Enclave Management
     * ==============================================================================================
     * Creates a new Kurtosis Enclave
     *
     * @generated from rpc engine_api.EngineService.CreateEnclave
     */
    createEnclave: {
      name: "CreateEnclave",
      I: CreateEnclaveArgs,
      O: CreateEnclaveResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Returns information about the existing enclaves
     *
     * @generated from rpc engine_api.EngineService.GetEnclaves
     */
    getEnclaves: {
      name: "GetEnclaves",
      I: Empty,
      O: GetEnclavesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Returns information about all existing & historical enclaves
     *
     * @generated from rpc engine_api.EngineService.GetExistingAndHistoricalEnclaveIdentifiers
     */
    getExistingAndHistoricalEnclaveIdentifiers: {
      name: "GetExistingAndHistoricalEnclaveIdentifiers",
      I: Empty,
      O: GetExistingAndHistoricalEnclaveIdentifiersResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Stops all containers in an enclave
     *
     * @generated from rpc engine_api.EngineService.StopEnclave
     */
    stopEnclave: {
      name: "StopEnclave",
      I: StopEnclaveArgs,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * Destroys an enclave, removing all artifacts associated with it
     *
     * @generated from rpc engine_api.EngineService.DestroyEnclave
     */
    destroyEnclave: {
      name: "DestroyEnclave",
      I: DestroyEnclaveArgs,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * Gets rid of old enclaves
     *
     * @generated from rpc engine_api.EngineService.Clean
     */
    clean: {
      name: "Clean",
      I: CleanArgs,
      O: CleanResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Get service logs
     *
     * @generated from rpc engine_api.EngineService.GetServiceLogs
     */
    getServiceLogs: {
      name: "GetServiceLogs",
      I: GetServiceLogsArgs,
      O: GetServiceLogsResponse,
      kind: MethodKind.ServerStreaming,
    },
  }
};


// @generated by protoc-gen-es v1.3.0 with parameter "target=js+dts"
// @generated from file engine_service.proto (package engine_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from enum engine_api.EnclaveMode
 */
export const EnclaveMode = proto3.makeEnum(
  "engine_api.EnclaveMode",
  [
    {no: 0, name: "TEST"},
    {no: 1, name: "PRODUCTION"},
  ],
);

/**
 * ==============================================================================================
 *                                            Get Enclaves
 * ==============================================================================================
 * Status of the containers in the enclave
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum valuee uniqueness works
 *
 * @generated from enum engine_api.EnclaveContainersStatus
 */
export const EnclaveContainersStatus = proto3.makeEnum(
  "engine_api.EnclaveContainersStatus",
  [
    {no: 0, name: "EnclaveContainersStatus_EMPTY"},
    {no: 1, name: "EnclaveContainersStatus_RUNNING"},
    {no: 2, name: "EnclaveContainersStatus_STOPPED"},
  ],
);

/**
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
 *
 * @generated from enum engine_api.EnclaveAPIContainerStatus
 */
export const EnclaveAPIContainerStatus = proto3.makeEnum(
  "engine_api.EnclaveAPIContainerStatus",
  [
    {no: 0, name: "EnclaveAPIContainerStatus_NONEXISTENT"},
    {no: 1, name: "EnclaveAPIContainerStatus_RUNNING"},
    {no: 2, name: "EnclaveAPIContainerStatus_STOPPED"},
  ],
);

/**
 * The filter operator which can be text or regex type
 * NOTE: We have to prefix the enum values with the enum name due to the way Protobuf enum value uniqueness works
 *
 * @generated from enum engine_api.LogLineOperator
 */
export const LogLineOperator = proto3.makeEnum(
  "engine_api.LogLineOperator",
  [
    {no: 0, name: "LogLineOperator_DOES_CONTAIN_TEXT"},
    {no: 1, name: "LogLineOperator_DOES_NOT_CONTAIN_TEXT"},
    {no: 2, name: "LogLineOperator_DOES_CONTAIN_MATCH_REGEX"},
    {no: 3, name: "LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX"},
  ],
);

/**
 * ==============================================================================================
 *                                        Get Engine Info
 * ==============================================================================================
 *
 * @generated from message engine_api.GetEngineInfoResponse
 */
export const GetEngineInfoResponse = proto3.makeMessageType(
  "engine_api.GetEngineInfoResponse",
  () => [
    { no: 1, name: "engine_version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                        Create Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.CreateEnclaveArgs
 */
export const CreateEnclaveArgs = proto3.makeMessageType(
  "engine_api.CreateEnclaveArgs",
  () => [
    { no: 1, name: "enclave_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "api_container_version_tag", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "api_container_log_level", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "mode", kind: "enum", T: proto3.getEnumType(EnclaveMode) },
  ],
);

/**
 * @generated from message engine_api.CreateEnclaveResponse
 */
export const CreateEnclaveResponse = proto3.makeMessageType(
  "engine_api.CreateEnclaveResponse",
  () => [
    { no: 1, name: "enclave_info", kind: "message", T: EnclaveInfo },
  ],
);

/**
 * @generated from message engine_api.EnclaveAPIContainerInfo
 */
export const EnclaveAPIContainerInfo = proto3.makeMessageType(
  "engine_api.EnclaveAPIContainerInfo",
  () => [
    { no: 1, name: "container_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "ip_inside_enclave", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "grpc_port_inside_enclave", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 6, name: "bridge_ip_address", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * Will only be present if the API container is running
 *
 * @generated from message engine_api.EnclaveAPIContainerHostMachineInfo
 */
export const EnclaveAPIContainerHostMachineInfo = proto3.makeMessageType(
  "engine_api.EnclaveAPIContainerHostMachineInfo",
  () => [
    { no: 4, name: "ip_on_host_machine", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "grpc_port_on_host_machine", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ],
);

/**
 * Enclaves are defined by a network in the container system, which is why there's a bunch of network information here
 *
 * @generated from message engine_api.EnclaveInfo
 */
export const EnclaveInfo = proto3.makeMessageType(
  "engine_api.EnclaveInfo",
  () => [
    { no: 1, name: "enclave_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "shortened_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "containers_status", kind: "enum", T: proto3.getEnumType(EnclaveContainersStatus) },
    { no: 5, name: "api_container_status", kind: "enum", T: proto3.getEnumType(EnclaveAPIContainerStatus) },
    { no: 6, name: "api_container_info", kind: "message", T: EnclaveAPIContainerInfo },
    { no: 7, name: "api_container_host_machine_info", kind: "message", T: EnclaveAPIContainerHostMachineInfo },
    { no: 8, name: "creation_time", kind: "message", T: Timestamp },
    { no: 9, name: "mode", kind: "enum", T: proto3.getEnumType(EnclaveMode) },
  ],
);

/**
 * @generated from message engine_api.GetEnclavesResponse
 */
export const GetEnclavesResponse = proto3.makeMessageType(
  "engine_api.GetEnclavesResponse",
  () => [
    { no: 1, name: "enclave_info", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: EnclaveInfo} },
  ],
);

/**
 * An enclave identifier is a collection of uuid, name and shortened uuid
 *
 * @generated from message engine_api.EnclaveIdentifiers
 */
export const EnclaveIdentifiers = proto3.makeMessageType(
  "engine_api.EnclaveIdentifiers",
  () => [
    { no: 1, name: "enclave_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "shortened_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message engine_api.GetExistingAndHistoricalEnclaveIdentifiersResponse
 */
export const GetExistingAndHistoricalEnclaveIdentifiersResponse = proto3.makeMessageType(
  "engine_api.GetExistingAndHistoricalEnclaveIdentifiersResponse",
  () => [
    { no: 1, name: "allIdentifiers", kind: "message", T: EnclaveIdentifiers, repeated: true },
  ],
);

/**
 * ==============================================================================================
 *                                       Stop Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.StopEnclaveArgs
 */
export const StopEnclaveArgs = proto3.makeMessageType(
  "engine_api.StopEnclaveArgs",
  () => [
    { no: 1, name: "enclave_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                       Destroy Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.DestroyEnclaveArgs
 */
export const DestroyEnclaveArgs = proto3.makeMessageType(
  "engine_api.DestroyEnclaveArgs",
  () => [
    { no: 1, name: "enclave_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                       Create Enclave
 * ==============================================================================================
 *
 * @generated from message engine_api.CleanArgs
 */
export const CleanArgs = proto3.makeMessageType(
  "engine_api.CleanArgs",
  () => [
    { no: 1, name: "should_clean_all", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ],
);

/**
 * @generated from message engine_api.EnclaveNameAndUuid
 */
export const EnclaveNameAndUuid = proto3.makeMessageType(
  "engine_api.EnclaveNameAndUuid",
  () => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message engine_api.CleanResponse
 */
export const CleanResponse = proto3.makeMessageType(
  "engine_api.CleanResponse",
  () => [
    { no: 1, name: "removed_enclave_name_and_uuids", kind: "message", T: EnclaveNameAndUuid, repeated: true },
  ],
);

/**
 * ==============================================================================================
 *                                   Get User Service Logs
 * ==============================================================================================
 *
 * @generated from message engine_api.GetServiceLogsArgs
 */
export const GetServiceLogsArgs = proto3.makeMessageType(
  "engine_api.GetServiceLogsArgs",
  () => [
    { no: 1, name: "enclave_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "service_uuid_set", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 8 /* ScalarType.BOOL */} },
    { no: 3, name: "follow_logs", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "conjunctive_filters", kind: "message", T: LogLineFilter, repeated: true },
    { no: 5, name: "return_all_logs", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 6, name: "num_log_lines", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ],
);

/**
 * @generated from message engine_api.GetServiceLogsResponse
 */
export const GetServiceLogsResponse = proto3.makeMessageType(
  "engine_api.GetServiceLogsResponse",
  () => [
    { no: 1, name: "service_logs_by_service_uuid", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: LogLine} },
    { no: 2, name: "not_found_service_uuid_set", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 8 /* ScalarType.BOOL */} },
  ],
);

/**
 * TODO add timestamp as well, for when we do timestamp-handling on the client side
 *
 * @generated from message engine_api.LogLine
 */
export const LogLine = proto3.makeMessageType(
  "engine_api.LogLine",
  () => [
    { no: 1, name: "line", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ],
);

/**
 * @generated from message engine_api.LogLineFilter
 */
export const LogLineFilter = proto3.makeMessageType(
  "engine_api.LogLineFilter",
  () => [
    { no: 1, name: "operator", kind: "enum", T: proto3.getEnumType(LogLineOperator) },
    { no: 2, name: "text_pattern", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);


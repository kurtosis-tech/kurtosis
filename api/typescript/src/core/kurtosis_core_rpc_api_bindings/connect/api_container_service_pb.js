// @generated by protoc-gen-es v1.3.1 with parameter "target=js+dts"
// @generated from file api_container_service.proto (package api_container_api, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum api_container_api.ServiceStatus
 */
export const ServiceStatus = proto3.makeEnum(
  "api_container_api.ServiceStatus",
  [
    {no: 0, name: "STOPPED"},
    {no: 1, name: "RUNNING"},
    {no: 2, name: "UNKNOWN"},
  ],
);

/**
 * @generated from enum api_container_api.ImageDownloadMode
 */
export const ImageDownloadMode = proto3.makeEnum(
  "api_container_api.ImageDownloadMode",
  [
    {no: 0, name: "always"},
    {no: 1, name: "missing"},
  ],
);

/**
 * User services port forwarding
 *
 * @generated from enum api_container_api.Connect
 */
export const Connect = proto3.makeEnum(
  "api_container_api.Connect",
  [
    {no: 0, name: "CONNECT"},
    {no: 1, name: "NO_CONNECT"},
  ],
);

/**
 * @generated from enum api_container_api.KurtosisFeatureFlag
 */
export const KurtosisFeatureFlag = proto3.makeEnum(
  "api_container_api.KurtosisFeatureFlag",
  [
    {no: 0, name: "NO_INSTRUCTIONS_CACHING"},
  ],
);

/**
 * @generated from enum api_container_api.RestartPolicy
 */
export const RestartPolicy = proto3.makeEnum(
  "api_container_api.RestartPolicy",
  [
    {no: 0, name: "NEVER"},
    {no: 1, name: "ALWAYS"},
  ],
);

/**
 * ==============================================================================================
 *                           Shared Objects (Used By Multiple Endpoints)
 * ==============================================================================================
 *
 * @generated from message api_container_api.Port
 */
export const Port = proto3.makeMessageType(
  "api_container_api.Port",
  () => [
    { no: 1, name: "number", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 2, name: "transport_protocol", kind: "enum", T: proto3.getEnumType(Port_TransportProtocol) },
    { no: 3, name: "maybe_application_protocol", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "maybe_wait_timeout", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "locked", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 6, name: "alias", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * @generated from enum api_container_api.Port.TransportProtocol
 */
export const Port_TransportProtocol = proto3.makeEnum(
  "api_container_api.Port.TransportProtocol",
  [
    {no: 0, name: "TCP"},
    {no: 1, name: "SCTP"},
    {no: 2, name: "UDP"},
  ],
);

/**
 * @generated from message api_container_api.Container
 */
export const Container = proto3.makeMessageType(
  "api_container_api.Container",
  () => [
    { no: 1, name: "status", kind: "enum", T: proto3.getEnumType(Container_Status) },
    { no: 2, name: "image_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "entrypoint_args", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "cmd_args", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 5, name: "env_vars", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
  ],
);

/**
 * @generated from enum api_container_api.Container.Status
 */
export const Container_Status = proto3.makeEnum(
  "api_container_api.Container.Status",
  [
    {no: 0, name: "STOPPED"},
    {no: 1, name: "RUNNING"},
    {no: 2, name: "UNKNOWN"},
  ],
);

/**
 * @generated from message api_container_api.FilesArtifactsList
 */
export const FilesArtifactsList = proto3.makeMessageType(
  "api_container_api.FilesArtifactsList",
  () => [
    { no: 1, name: "files_artifacts_identifiers", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ],
);

/**
 * Equivalent of user on ServiceConfig
 *
 * @generated from message api_container_api.User
 */
export const User = proto3.makeMessageType(
  "api_container_api.User",
  () => [
    { no: 1, name: "uid", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 2, name: "gid", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ],
);

/**
 * Equivalent of tolerations on ServiceConfig
 *
 * @generated from message api_container_api.Toleration
 */
export const Toleration = proto3.makeMessageType(
  "api_container_api.Toleration",
  () => [
    { no: 1, name: "key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "operator", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "value", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "effect", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "toleration_seconds", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ],
);

/**
 * @generated from message api_container_api.ServiceInfo
 */
export const ServiceInfo = proto3.makeMessageType(
  "api_container_api.ServiceInfo",
  () => [
    { no: 1, name: "service_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "private_ip_addr", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "private_ports", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: Port} },
    { no: 4, name: "maybe_public_ip_addr", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "maybe_public_ports", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: Port} },
    { no: 6, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 7, name: "shortened_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 8, name: "service_status", kind: "enum", T: proto3.getEnumType(ServiceStatus) },
    { no: 9, name: "container", kind: "message", T: Container },
    { no: 10, name: "service_dir_paths_to_files_artifacts_list", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: FilesArtifactsList} },
    { no: 11, name: "max_millicpus", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 12, name: "min_millicpus", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 13, name: "max_memory_megabytes", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 14, name: "min_memory_megabytes", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 15, name: "user", kind: "message", T: User, opt: true },
    { no: 16, name: "tolerations", kind: "message", T: Toleration, repeated: true },
    { no: 17, name: "node_selectors", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
    { no: 18, name: "labels", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 9 /* ScalarType.STRING */} },
    { no: 19, name: "tini_enabled", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
  ],
);

/**
 * @generated from message api_container_api.RunStarlarkScriptArgs
 */
export const RunStarlarkScriptArgs = proto3.makeMessageType(
  "api_container_api.RunStarlarkScriptArgs",
  () => [
    { no: 1, name: "serialized_script", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "dry_run", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 4, name: "parallelism", kind: "scalar", T: 5 /* ScalarType.INT32 */, opt: true },
    { no: 5, name: "main_function_name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 6, name: "experimental_features", kind: "enum", T: proto3.getEnumType(KurtosisFeatureFlag), repeated: true },
    { no: 7, name: "cloud_instance_id", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 8, name: "cloud_user_id", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 9, name: "image_download_mode", kind: "enum", T: proto3.getEnumType(ImageDownloadMode), opt: true },
    { no: 10, name: "non_blocking_mode", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
  ],
);

/**
 * @generated from message api_container_api.RunStarlarkPackageArgs
 */
export const RunStarlarkPackageArgs = proto3.makeMessageType(
  "api_container_api.RunStarlarkPackageArgs",
  () => [
    { no: 1, name: "package_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "local", kind: "scalar", T: 12 /* ScalarType.BYTES */, oneof: "starlark_package_content" },
    { no: 4, name: "remote", kind: "scalar", T: 8 /* ScalarType.BOOL */, oneof: "starlark_package_content" },
    { no: 5, name: "serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 6, name: "dry_run", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 7, name: "parallelism", kind: "scalar", T: 5 /* ScalarType.INT32 */, opt: true },
    { no: 8, name: "clone_package", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 9, name: "relative_path_to_main_file", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 10, name: "main_function_name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 11, name: "experimental_features", kind: "enum", T: proto3.getEnumType(KurtosisFeatureFlag), repeated: true },
    { no: 12, name: "cloud_instance_id", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 13, name: "cloud_user_id", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 14, name: "image_download_mode", kind: "enum", T: proto3.getEnumType(ImageDownloadMode), opt: true },
    { no: 15, name: "non_blocking_mode", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 16, name: "github_auth_token", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * ==============================================================================================
 *                               Starlark Execution Response
 * ==============================================================================================
 *
 * @generated from message api_container_api.StarlarkRunResponseLine
 */
export const StarlarkRunResponseLine = proto3.makeMessageType(
  "api_container_api.StarlarkRunResponseLine",
  () => [
    { no: 1, name: "instruction", kind: "message", T: StarlarkInstruction, oneof: "run_response_line" },
    { no: 2, name: "error", kind: "message", T: StarlarkError, oneof: "run_response_line" },
    { no: 3, name: "progress_info", kind: "message", T: StarlarkRunProgress, oneof: "run_response_line" },
    { no: 4, name: "instruction_result", kind: "message", T: StarlarkInstructionResult, oneof: "run_response_line" },
    { no: 5, name: "run_finished_event", kind: "message", T: StarlarkRunFinishedEvent, oneof: "run_response_line" },
    { no: 6, name: "warning", kind: "message", T: StarlarkWarning, oneof: "run_response_line" },
    { no: 7, name: "info", kind: "message", T: StarlarkInfo, oneof: "run_response_line" },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInfo
 */
export const StarlarkInfo = proto3.makeMessageType(
  "api_container_api.StarlarkInfo",
  () => [
    { no: 1, name: "info_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkWarning
 */
export const StarlarkWarning = proto3.makeMessageType(
  "api_container_api.StarlarkWarning",
  () => [
    { no: 1, name: "warning_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInstruction
 */
export const StarlarkInstruction = proto3.makeMessageType(
  "api_container_api.StarlarkInstruction",
  () => [
    { no: 1, name: "position", kind: "message", T: StarlarkInstructionPosition },
    { no: 2, name: "instruction_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "arguments", kind: "message", T: StarlarkInstructionArg, repeated: true },
    { no: 4, name: "executable_instruction", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "is_skipped", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 6, name: "description", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInstructionResult
 */
export const StarlarkInstructionResult = proto3.makeMessageType(
  "api_container_api.StarlarkInstructionResult",
  () => [
    { no: 1, name: "serialized_instruction_result", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInstructionArg
 */
export const StarlarkInstructionArg = proto3.makeMessageType(
  "api_container_api.StarlarkInstructionArg",
  () => [
    { no: 1, name: "serialized_arg_value", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "arg_name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "is_representative", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInstructionPosition
 */
export const StarlarkInstructionPosition = proto3.makeMessageType(
  "api_container_api.StarlarkInstructionPosition",
  () => [
    { no: 1, name: "filename", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "line", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "column", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkError
 */
export const StarlarkError = proto3.makeMessageType(
  "api_container_api.StarlarkError",
  () => [
    { no: 1, name: "interpretation_error", kind: "message", T: StarlarkInterpretationError, oneof: "error" },
    { no: 2, name: "validation_error", kind: "message", T: StarlarkValidationError, oneof: "error" },
    { no: 3, name: "execution_error", kind: "message", T: StarlarkExecutionError, oneof: "error" },
  ],
);

/**
 * @generated from message api_container_api.StarlarkInterpretationError
 */
export const StarlarkInterpretationError = proto3.makeMessageType(
  "api_container_api.StarlarkInterpretationError",
  () => [
    { no: 1, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkValidationError
 */
export const StarlarkValidationError = proto3.makeMessageType(
  "api_container_api.StarlarkValidationError",
  () => [
    { no: 1, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkExecutionError
 */
export const StarlarkExecutionError = proto3.makeMessageType(
  "api_container_api.StarlarkExecutionError",
  () => [
    { no: 1, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkRunProgress
 */
export const StarlarkRunProgress = proto3.makeMessageType(
  "api_container_api.StarlarkRunProgress",
  () => [
    { no: 1, name: "current_step_info", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 2, name: "total_steps", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 3, name: "current_step_number", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkRunFinishedEvent
 */
export const StarlarkRunFinishedEvent = proto3.makeMessageType(
  "api_container_api.StarlarkRunFinishedEvent",
  () => [
    { no: 1, name: "is_run_successful", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 2, name: "serialized_output", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * ==============================================================================================
 *                                          Get Services
 * ==============================================================================================
 *
 * @generated from message api_container_api.GetServicesArgs
 */
export const GetServicesArgs = proto3.makeMessageType(
  "api_container_api.GetServicesArgs",
  () => [
    { no: 1, name: "service_identifiers", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "scalar", T: 8 /* ScalarType.BOOL */} },
  ],
);

/**
 * @generated from message api_container_api.GetServicesResponse
 */
export const GetServicesResponse = proto3.makeMessageType(
  "api_container_api.GetServicesResponse",
  () => [
    { no: 1, name: "service_info", kind: "map", K: 9 /* ScalarType.STRING */, V: {kind: "message", T: ServiceInfo} },
  ],
);

/**
 * An service identifier is a collection of uuid, name and shortened uuid
 *
 * @generated from message api_container_api.ServiceIdentifiers
 */
export const ServiceIdentifiers = proto3.makeMessageType(
  "api_container_api.ServiceIdentifiers",
  () => [
    { no: 1, name: "service_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "shortened_uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse
 */
export const GetExistingAndHistoricalServiceIdentifiersResponse = proto3.makeMessageType(
  "api_container_api.GetExistingAndHistoricalServiceIdentifiersResponse",
  () => [
    { no: 1, name: "allIdentifiers", kind: "message", T: ServiceIdentifiers, repeated: true },
  ],
);

/**
 * ==============================================================================================
 *                                          Exec Command
 * ==============================================================================================
 *
 * @generated from message api_container_api.ExecCommandArgs
 */
export const ExecCommandArgs = proto3.makeMessageType(
  "api_container_api.ExecCommandArgs",
  () => [
    { no: 1, name: "service_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "command_args", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ],
);

/**
 * @generated from message api_container_api.ExecCommandResponse
 */
export const ExecCommandResponse = proto3.makeMessageType(
  "api_container_api.ExecCommandResponse",
  () => [
    { no: 1, name: "exit_code", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 2, name: "log_output", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                             Wait For HTTP Get Endpoint Availability
 * ==============================================================================================
 *
 * @generated from message api_container_api.WaitForHttpGetEndpointAvailabilityArgs
 */
export const WaitForHttpGetEndpointAvailabilityArgs = proto3.makeMessageType(
  "api_container_api.WaitForHttpGetEndpointAvailabilityArgs",
  () => [
    { no: 1, name: "service_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "port", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 3, name: "path", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 4, name: "initial_delay_milliseconds", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 5, name: "retries", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 6, name: "retries_delay_milliseconds", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 7, name: "body_text", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * ==============================================================================================
 *                           Wait For HTTP Post Endpoint Availability
 * ==============================================================================================
 *
 * @generated from message api_container_api.WaitForHttpPostEndpointAvailabilityArgs
 */
export const WaitForHttpPostEndpointAvailabilityArgs = proto3.makeMessageType(
  "api_container_api.WaitForHttpPostEndpointAvailabilityArgs",
  () => [
    { no: 1, name: "service_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "port", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
    { no: 3, name: "path", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 4, name: "request_body", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 5, name: "initial_delay_milliseconds", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 6, name: "retries", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 7, name: "retries_delay_milliseconds", kind: "scalar", T: 13 /* ScalarType.UINT32 */, opt: true },
    { no: 8, name: "body_text", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * ==============================================================================================
 *                                          Streamed Data Chunk
 * ==============================================================================================
 *
 * @generated from message api_container_api.StreamedDataChunk
 */
export const StreamedDataChunk = proto3.makeMessageType(
  "api_container_api.StreamedDataChunk",
  () => [
    { no: 1, name: "data", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
    { no: 2, name: "previous_chunk_hash", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "metadata", kind: "message", T: DataChunkMetadata },
  ],
);

/**
 * @generated from message api_container_api.DataChunkMetadata
 */
export const DataChunkMetadata = proto3.makeMessageType(
  "api_container_api.DataChunkMetadata",
  () => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                          Upload Files Artifact
 * ==============================================================================================
 *
 * @generated from message api_container_api.UploadFilesArtifactResponse
 */
export const UploadFilesArtifactResponse = proto3.makeMessageType(
  "api_container_api.UploadFilesArtifactResponse",
  () => [
    { no: 1, name: "uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                          Download Files Artifact
 * ==============================================================================================
 *
 * @generated from message api_container_api.DownloadFilesArtifactArgs
 */
export const DownloadFilesArtifactArgs = proto3.makeMessageType(
  "api_container_api.DownloadFilesArtifactArgs",
  () => [
    { no: 1, name: "identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * ==============================================================================================
 *                                        Store Web Files Artifact
 * ==============================================================================================
 *
 * @generated from message api_container_api.StoreWebFilesArtifactArgs
 */
export const StoreWebFilesArtifactArgs = proto3.makeMessageType(
  "api_container_api.StoreWebFilesArtifactArgs",
  () => [
    { no: 1, name: "url", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StoreWebFilesArtifactResponse
 */
export const StoreWebFilesArtifactResponse = proto3.makeMessageType(
  "api_container_api.StoreWebFilesArtifactResponse",
  () => [
    { no: 1, name: "uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StoreFilesArtifactFromServiceArgs
 */
export const StoreFilesArtifactFromServiceArgs = proto3.makeMessageType(
  "api_container_api.StoreFilesArtifactFromServiceArgs",
  () => [
    { no: 1, name: "service_identifier", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "source_path", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StoreFilesArtifactFromServiceResponse
 */
export const StoreFilesArtifactFromServiceResponse = proto3.makeMessageType(
  "api_container_api.StoreFilesArtifactFromServiceResponse",
  () => [
    { no: 1, name: "uuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.FilesArtifactNameAndUuid
 */
export const FilesArtifactNameAndUuid = proto3.makeMessageType(
  "api_container_api.FilesArtifactNameAndUuid",
  () => [
    { no: 1, name: "fileName", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "fileUuid", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.ListFilesArtifactNamesAndUuidsResponse
 */
export const ListFilesArtifactNamesAndUuidsResponse = proto3.makeMessageType(
  "api_container_api.ListFilesArtifactNamesAndUuidsResponse",
  () => [
    { no: 1, name: "file_names_and_uuids", kind: "message", T: FilesArtifactNameAndUuid, repeated: true },
  ],
);

/**
 * @generated from message api_container_api.InspectFilesArtifactContentsRequest
 */
export const InspectFilesArtifactContentsRequest = proto3.makeMessageType(
  "api_container_api.InspectFilesArtifactContentsRequest",
  () => [
    { no: 1, name: "file_names_and_uuid", kind: "message", T: FilesArtifactNameAndUuid },
  ],
);

/**
 * @generated from message api_container_api.InspectFilesArtifactContentsResponse
 */
export const InspectFilesArtifactContentsResponse = proto3.makeMessageType(
  "api_container_api.InspectFilesArtifactContentsResponse",
  () => [
    { no: 1, name: "file_descriptions", kind: "message", T: FileArtifactContentsFileDescription, repeated: true },
  ],
);

/**
 * @generated from message api_container_api.FileArtifactContentsFileDescription
 */
export const FileArtifactContentsFileDescription = proto3.makeMessageType(
  "api_container_api.FileArtifactContentsFileDescription",
  () => [
    { no: 1, name: "path", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "size", kind: "scalar", T: 4 /* ScalarType.UINT64 */ },
    { no: 3, name: "text_preview", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * @generated from message api_container_api.ConnectServicesArgs
 */
export const ConnectServicesArgs = proto3.makeMessageType(
  "api_container_api.ConnectServicesArgs",
  () => [
    { no: 1, name: "connect", kind: "enum", T: proto3.getEnumType(Connect) },
  ],
);

/**
 * @generated from message api_container_api.ConnectServicesResponse
 */
export const ConnectServicesResponse = proto3.makeMessageType(
  "api_container_api.ConnectServicesResponse",
  [],
);

/**
 * @generated from message api_container_api.GetStarlarkRunResponse
 */
export const GetStarlarkRunResponse = proto3.makeMessageType(
  "api_container_api.GetStarlarkRunResponse",
  () => [
    { no: 1, name: "package_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "serialized_script", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "parallelism", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 5, name: "relative_path_to_main_file", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 6, name: "main_function_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 7, name: "experimental_features", kind: "enum", T: proto3.getEnumType(KurtosisFeatureFlag), repeated: true },
    { no: 8, name: "restart_policy", kind: "enum", T: proto3.getEnumType(RestartPolicy) },
    { no: 9, name: "initial_serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * @generated from message api_container_api.PlanYaml
 */
export const PlanYaml = proto3.makeMessageType(
  "api_container_api.PlanYaml",
  () => [
    { no: 1, name: "plan_yaml", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ],
);

/**
 * @generated from message api_container_api.StarlarkScriptPlanYamlArgs
 */
export const StarlarkScriptPlanYamlArgs = proto3.makeMessageType(
  "api_container_api.StarlarkScriptPlanYamlArgs",
  () => [
    { no: 1, name: "serialized_script", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 5, name: "main_function_name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);

/**
 * @generated from message api_container_api.StarlarkPackagePlanYamlArgs
 */
export const StarlarkPackagePlanYamlArgs = proto3.makeMessageType(
  "api_container_api.StarlarkPackagePlanYamlArgs",
  () => [
    { no: 1, name: "package_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "serialized_params", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "is_remote", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "relative_path_to_main_file", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 5, name: "main_function_name", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ],
);


defmodule ApiContainerApi.ServiceStatus do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :STOPPED, 0
  field :RUNNING, 1
  field :UNKNOWN, 2
end

defmodule ApiContainerApi.Connect do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :CONNECT, 0
  field :NO_CONNECT, 1
end

defmodule ApiContainerApi.KurtosisFeatureFlag do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :NO_INSTRUCTIONS_CACHING, 0
end

defmodule ApiContainerApi.RestartPolicy do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :NEVER, 0
  field :ALWAYS, 1
end

defmodule ApiContainerApi.Port.TransportProtocol do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :TCP, 0
  field :SCTP, 1
  field :UDP, 2
end

defmodule ApiContainerApi.Container.Status do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :STOPPED, 0
  field :RUNNING, 1
  field :UNKNOWN, 2
end

defmodule ApiContainerApi.Port do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :number, 1, type: :uint32

  field :transport_protocol, 2,
    type: ApiContainerApi.Port.TransportProtocol,
    json_name: "transportProtocol",
    enum: true

  field :maybe_application_protocol, 3, type: :string, json_name: "maybeApplicationProtocol"
  field :maybe_wait_timeout, 4, type: :string, json_name: "maybeWaitTimeout"
end

defmodule ApiContainerApi.Container.EnvVarsEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: :string
end

defmodule ApiContainerApi.Container do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :status, 1, type: ApiContainerApi.Container.Status, enum: true
  field :image_name, 2, type: :string, json_name: "imageName"
  field :entrypoint_args, 3, repeated: true, type: :string, json_name: "entrypointArgs"
  field :cmd_args, 4, repeated: true, type: :string, json_name: "cmdArgs"

  field :env_vars, 5,
    repeated: true,
    type: ApiContainerApi.Container.EnvVarsEntry,
    json_name: "envVars",
    map: true
end

defmodule ApiContainerApi.ServiceInfo.PrivatePortsEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: ApiContainerApi.Port
end

defmodule ApiContainerApi.ServiceInfo.MaybePublicPortsEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: ApiContainerApi.Port
end

defmodule ApiContainerApi.ServiceInfo do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_uuid, 1, type: :string, json_name: "serviceUuid"
  field :private_ip_addr, 2, type: :string, json_name: "privateIpAddr"

  field :private_ports, 3,
    repeated: true,
    type: ApiContainerApi.ServiceInfo.PrivatePortsEntry,
    json_name: "privatePorts",
    map: true

  field :maybe_public_ip_addr, 4, type: :string, json_name: "maybePublicIpAddr"

  field :maybe_public_ports, 5,
    repeated: true,
    type: ApiContainerApi.ServiceInfo.MaybePublicPortsEntry,
    json_name: "maybePublicPorts",
    map: true

  field :name, 6, type: :string
  field :shortened_uuid, 7, type: :string, json_name: "shortenedUuid"

  field :service_status, 8,
    type: ApiContainerApi.ServiceStatus,
    json_name: "serviceStatus",
    enum: true

  field :container, 9, type: ApiContainerApi.Container
end

defmodule ApiContainerApi.RunStarlarkScriptArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :serialized_script, 1, type: :string, json_name: "serializedScript"
  field :serialized_params, 2, type: :string, json_name: "serializedParams"
  field :dry_run, 3, proto3_optional: true, type: :bool, json_name: "dryRun"
  field :parallelism, 4, proto3_optional: true, type: :int32
  field :main_function_name, 5, type: :string, json_name: "mainFunctionName"

  field :experimental_features, 6,
    repeated: true,
    type: ApiContainerApi.KurtosisFeatureFlag,
    json_name: "experimentalFeatures",
    enum: true

  field :cloud_instance_id, 7, proto3_optional: true, type: :string, json_name: "cloudInstanceId"
  field :cloud_user_id, 8, proto3_optional: true, type: :string, json_name: "cloudUserId"
end

defmodule ApiContainerApi.RunStarlarkPackageArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  oneof :starlark_package_content, 0

  field :package_id, 1, type: :string, json_name: "packageId"
  field :local, 3, type: :bytes, oneof: 0
  field :remote, 4, type: :bool, oneof: 0
  field :serialized_params, 5, type: :string, json_name: "serializedParams"
  field :dry_run, 6, proto3_optional: true, type: :bool, json_name: "dryRun"
  field :parallelism, 7, proto3_optional: true, type: :int32
  field :clone_package, 8, proto3_optional: true, type: :bool, json_name: "clonePackage"
  field :relative_path_to_main_file, 9, type: :string, json_name: "relativePathToMainFile"
  field :main_function_name, 10, type: :string, json_name: "mainFunctionName"

  field :experimental_features, 11,
    repeated: true,
    type: ApiContainerApi.KurtosisFeatureFlag,
    json_name: "experimentalFeatures",
    enum: true

  field :cloud_instance_id, 12, proto3_optional: true, type: :string, json_name: "cloudInstanceId"
  field :cloud_user_id, 13, proto3_optional: true, type: :string, json_name: "cloudUserId"
end

defmodule ApiContainerApi.StarlarkRunResponseLine do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  oneof :run_response_line, 0

  field :instruction, 1, type: ApiContainerApi.StarlarkInstruction, oneof: 0
  field :error, 2, type: ApiContainerApi.StarlarkError, oneof: 0

  field :progress_info, 3,
    type: ApiContainerApi.StarlarkRunProgress,
    json_name: "progressInfo",
    oneof: 0

  field :instruction_result, 4,
    type: ApiContainerApi.StarlarkInstructionResult,
    json_name: "instructionResult",
    oneof: 0

  field :run_finished_event, 5,
    type: ApiContainerApi.StarlarkRunFinishedEvent,
    json_name: "runFinishedEvent",
    oneof: 0

  field :warning, 6, type: ApiContainerApi.StarlarkWarning, oneof: 0
  field :info, 7, type: ApiContainerApi.StarlarkInfo, oneof: 0
end

defmodule ApiContainerApi.StarlarkInfo do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :info_message, 1, type: :string, json_name: "infoMessage"
end

defmodule ApiContainerApi.StarlarkWarning do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :warning_message, 1, type: :string, json_name: "warningMessage"
end

defmodule ApiContainerApi.StarlarkInstruction do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :position, 1, type: ApiContainerApi.StarlarkInstructionPosition
  field :instruction_name, 2, type: :string, json_name: "instructionName"
  field :arguments, 3, repeated: true, type: ApiContainerApi.StarlarkInstructionArg
  field :executable_instruction, 4, type: :string, json_name: "executableInstruction"
  field :is_skipped, 5, type: :bool, json_name: "isSkipped"
end

defmodule ApiContainerApi.StarlarkInstructionResult do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :serialized_instruction_result, 1, type: :string, json_name: "serializedInstructionResult"
end

defmodule ApiContainerApi.StarlarkInstructionArg do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :serialized_arg_value, 1, type: :string, json_name: "serializedArgValue"
  field :arg_name, 2, proto3_optional: true, type: :string, json_name: "argName"
  field :is_representative, 3, type: :bool, json_name: "isRepresentative"
end

defmodule ApiContainerApi.StarlarkInstructionPosition do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :filename, 1, type: :string
  field :line, 2, type: :int32
  field :column, 3, type: :int32
end

defmodule ApiContainerApi.StarlarkError do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  oneof :error, 0

  field :interpretation_error, 1,
    type: ApiContainerApi.StarlarkInterpretationError,
    json_name: "interpretationError",
    oneof: 0

  field :validation_error, 2,
    type: ApiContainerApi.StarlarkValidationError,
    json_name: "validationError",
    oneof: 0

  field :execution_error, 3,
    type: ApiContainerApi.StarlarkExecutionError,
    json_name: "executionError",
    oneof: 0
end

defmodule ApiContainerApi.StarlarkInterpretationError do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :error_message, 1, type: :string, json_name: "errorMessage"
end

defmodule ApiContainerApi.StarlarkValidationError do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :error_message, 1, type: :string, json_name: "errorMessage"
end

defmodule ApiContainerApi.StarlarkExecutionError do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :error_message, 1, type: :string, json_name: "errorMessage"
end

defmodule ApiContainerApi.StarlarkRunProgress do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :current_step_info, 1, repeated: true, type: :string, json_name: "currentStepInfo"
  field :total_steps, 2, type: :uint32, json_name: "totalSteps"
  field :current_step_number, 3, type: :uint32, json_name: "currentStepNumber"
end

defmodule ApiContainerApi.StarlarkRunFinishedEvent do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :is_run_successful, 1, type: :bool, json_name: "isRunSuccessful"
  field :serialized_output, 2, proto3_optional: true, type: :string, json_name: "serializedOutput"
end

defmodule ApiContainerApi.GetServicesArgs.ServiceIdentifiersEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: :bool
end

defmodule ApiContainerApi.GetServicesArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_identifiers, 1,
    repeated: true,
    type: ApiContainerApi.GetServicesArgs.ServiceIdentifiersEntry,
    json_name: "serviceIdentifiers",
    map: true
end

defmodule ApiContainerApi.GetServicesResponse.ServiceInfoEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: ApiContainerApi.ServiceInfo
end

defmodule ApiContainerApi.GetServicesResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_info, 1,
    repeated: true,
    type: ApiContainerApi.GetServicesResponse.ServiceInfoEntry,
    json_name: "serviceInfo",
    map: true
end

defmodule ApiContainerApi.ServiceIdentifiers do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_uuid, 1, type: :string, json_name: "serviceUuid"
  field :name, 2, type: :string
  field :shortened_uuid, 3, type: :string, json_name: "shortenedUuid"
end

defmodule ApiContainerApi.GetExistingAndHistoricalServiceIdentifiersResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :allIdentifiers, 1, repeated: true, type: ApiContainerApi.ServiceIdentifiers
end

defmodule ApiContainerApi.ExecCommandArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_identifier, 1, type: :string, json_name: "serviceIdentifier"
  field :command_args, 2, repeated: true, type: :string, json_name: "commandArgs"
end

defmodule ApiContainerApi.ExecCommandResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :exit_code, 1, type: :int32, json_name: "exitCode"
  field :log_output, 2, type: :string, json_name: "logOutput"
end

defmodule ApiContainerApi.WaitForHttpGetEndpointAvailabilityArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_identifier, 1, type: :string, json_name: "serviceIdentifier"
  field :port, 2, type: :uint32
  field :path, 3, type: :string
  field :initial_delay_milliseconds, 4, type: :uint32, json_name: "initialDelayMilliseconds"
  field :retries, 5, type: :uint32
  field :retries_delay_milliseconds, 6, type: :uint32, json_name: "retriesDelayMilliseconds"
  field :body_text, 7, type: :string, json_name: "bodyText"
end

defmodule ApiContainerApi.WaitForHttpPostEndpointAvailabilityArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_identifier, 1, type: :string, json_name: "serviceIdentifier"
  field :port, 2, type: :uint32
  field :path, 3, type: :string
  field :request_body, 4, type: :string, json_name: "requestBody"
  field :initial_delay_milliseconds, 5, type: :uint32, json_name: "initialDelayMilliseconds"
  field :retries, 6, type: :uint32
  field :retries_delay_milliseconds, 7, type: :uint32, json_name: "retriesDelayMilliseconds"
  field :body_text, 8, type: :string, json_name: "bodyText"
end

defmodule ApiContainerApi.StreamedDataChunk do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :data, 1, type: :bytes
  field :previous_chunk_hash, 2, type: :string, json_name: "previousChunkHash"
  field :metadata, 3, type: ApiContainerApi.DataChunkMetadata
end

defmodule ApiContainerApi.DataChunkMetadata do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :name, 1, type: :string
end

defmodule ApiContainerApi.UploadFilesArtifactResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :uuid, 1, type: :string
  field :name, 2, type: :string
end

defmodule ApiContainerApi.DownloadFilesArtifactArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :identifier, 1, type: :string
end

defmodule ApiContainerApi.StoreWebFilesArtifactArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :url, 1, type: :string
  field :name, 2, type: :string
end

defmodule ApiContainerApi.StoreWebFilesArtifactResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :uuid, 1, type: :string
end

defmodule ApiContainerApi.StoreFilesArtifactFromServiceArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_identifier, 1, type: :string, json_name: "serviceIdentifier"
  field :source_path, 2, type: :string, json_name: "sourcePath"
  field :name, 3, type: :string
end

defmodule ApiContainerApi.StoreFilesArtifactFromServiceResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :uuid, 1, type: :string
end

defmodule ApiContainerApi.FilesArtifactNameAndUuid do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :fileName, 1, type: :string
  field :fileUuid, 2, type: :string
end

defmodule ApiContainerApi.ListFilesArtifactNamesAndUuidsResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :file_names_and_uuids, 1,
    repeated: true,
    type: ApiContainerApi.FilesArtifactNameAndUuid,
    json_name: "fileNamesAndUuids"
end

defmodule ApiContainerApi.InspectFilesArtifactContentsRequest do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :file_names_and_uuid, 1,
    type: ApiContainerApi.FilesArtifactNameAndUuid,
    json_name: "fileNamesAndUuid"
end

defmodule ApiContainerApi.InspectFilesArtifactContentsResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :file_descriptions, 1,
    repeated: true,
    type: ApiContainerApi.FileArtifactContentsFileDescription,
    json_name: "fileDescriptions"
end

defmodule ApiContainerApi.FileArtifactContentsFileDescription do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :path, 1, type: :string
  field :size, 2, type: :uint64
  field :text_preview, 3, proto3_optional: true, type: :string, json_name: "textPreview"
end

defmodule ApiContainerApi.ConnectServicesArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :connect, 1, type: ApiContainerApi.Connect, enum: true
end

defmodule ApiContainerApi.ConnectServicesResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3
end

defmodule ApiContainerApi.GetStarlarkRunResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :package_id, 1, type: :string, json_name: "packageId"
  field :serialized_script, 2, type: :string, json_name: "serializedScript"
  field :serialized_params, 3, type: :string, json_name: "serializedParams"
  field :parallelism, 4, type: :int32
  field :relative_path_to_main_file, 5, type: :string, json_name: "relativePathToMainFile"
  field :main_function_name, 6, type: :string, json_name: "mainFunctionName"

  field :experimental_features, 7,
    repeated: true,
    type: ApiContainerApi.KurtosisFeatureFlag,
    json_name: "experimentalFeatures",
    enum: true

  field :restart_policy, 8,
    type: ApiContainerApi.RestartPolicy,
    json_name: "restartPolicy",
    enum: true
end

defmodule ApiContainerApi.ApiContainerService.Service do
  @moduledoc false

  use GRPC.Service,
    name: "api_container_api.ApiContainerService",
    protoc_gen_elixir_version: "0.12.0"

  rpc :RunStarlarkScript,
      ApiContainerApi.RunStarlarkScriptArgs,
      stream(ApiContainerApi.StarlarkRunResponseLine)

  rpc :UploadStarlarkPackage, stream(ApiContainerApi.StreamedDataChunk), Google.Protobuf.Empty

  rpc :RunStarlarkPackage,
      ApiContainerApi.RunStarlarkPackageArgs,
      stream(ApiContainerApi.StarlarkRunResponseLine)

  rpc :GetServices, ApiContainerApi.GetServicesArgs, ApiContainerApi.GetServicesResponse

  rpc :GetExistingAndHistoricalServiceIdentifiers,
      Google.Protobuf.Empty,
      ApiContainerApi.GetExistingAndHistoricalServiceIdentifiersResponse

  rpc :ExecCommand, ApiContainerApi.ExecCommandArgs, ApiContainerApi.ExecCommandResponse

  rpc :WaitForHttpGetEndpointAvailability,
      ApiContainerApi.WaitForHttpGetEndpointAvailabilityArgs,
      Google.Protobuf.Empty

  rpc :WaitForHttpPostEndpointAvailability,
      ApiContainerApi.WaitForHttpPostEndpointAvailabilityArgs,
      Google.Protobuf.Empty

  rpc :UploadFilesArtifact,
      stream(ApiContainerApi.StreamedDataChunk),
      ApiContainerApi.UploadFilesArtifactResponse

  rpc :DownloadFilesArtifact,
      ApiContainerApi.DownloadFilesArtifactArgs,
      stream(ApiContainerApi.StreamedDataChunk)

  rpc :StoreWebFilesArtifact,
      ApiContainerApi.StoreWebFilesArtifactArgs,
      ApiContainerApi.StoreWebFilesArtifactResponse

  rpc :StoreFilesArtifactFromService,
      ApiContainerApi.StoreFilesArtifactFromServiceArgs,
      ApiContainerApi.StoreFilesArtifactFromServiceResponse

  rpc :ListFilesArtifactNamesAndUuids,
      Google.Protobuf.Empty,
      ApiContainerApi.ListFilesArtifactNamesAndUuidsResponse

  rpc :InspectFilesArtifactContents,
      ApiContainerApi.InspectFilesArtifactContentsRequest,
      ApiContainerApi.InspectFilesArtifactContentsResponse

  rpc :ConnectServices,
      ApiContainerApi.ConnectServicesArgs,
      ApiContainerApi.ConnectServicesResponse

  rpc :GetStarlarkRun, Google.Protobuf.Empty, ApiContainerApi.GetStarlarkRunResponse
end

defmodule ApiContainerApi.ApiContainerService.Stub do
  @moduledoc false

  use GRPC.Stub, service: ApiContainerApi.ApiContainerService.Service
end

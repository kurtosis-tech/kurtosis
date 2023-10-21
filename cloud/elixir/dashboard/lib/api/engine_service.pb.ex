defmodule EngineApi.EnclaveMode do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :TEST, 0
  field :PRODUCTION, 1
end

defmodule EngineApi.EnclaveContainersStatus do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :EnclaveContainersStatus_EMPTY, 0
  field :EnclaveContainersStatus_RUNNING, 1
  field :EnclaveContainersStatus_STOPPED, 2
end

defmodule EngineApi.EnclaveAPIContainerStatus do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :EnclaveAPIContainerStatus_NONEXISTENT, 0
  field :EnclaveAPIContainerStatus_RUNNING, 1
  field :EnclaveAPIContainerStatus_STOPPED, 2
end

defmodule EngineApi.LogLineOperator do
  @moduledoc false

  use Protobuf, enum: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :LogLineOperator_DOES_CONTAIN_TEXT, 0
  field :LogLineOperator_DOES_NOT_CONTAIN_TEXT, 1
  field :LogLineOperator_DOES_CONTAIN_MATCH_REGEX, 2
  field :LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX, 3
end

defmodule EngineApi.GetEngineInfoResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :engine_version, 1, type: :string, json_name: "engineVersion"
end

defmodule EngineApi.CreateEnclaveArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_name, 1, type: :string, json_name: "enclaveName"
  field :api_container_version_tag, 2, type: :string, json_name: "apiContainerVersionTag"
  field :api_container_log_level, 3, type: :string, json_name: "apiContainerLogLevel"
  field :mode, 4, type: EngineApi.EnclaveMode, enum: true
end

defmodule EngineApi.CreateEnclaveResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_info, 1, type: EngineApi.EnclaveInfo, json_name: "enclaveInfo"
end

defmodule EngineApi.EnclaveAPIContainerInfo do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :container_id, 1, type: :string, json_name: "containerId"
  field :ip_inside_enclave, 2, type: :string, json_name: "ipInsideEnclave"
  field :grpc_port_inside_enclave, 3, type: :uint32, json_name: "grpcPortInsideEnclave"
  field :bridge_ip_address, 6, type: :string, json_name: "bridgeIpAddress"
end

defmodule EngineApi.EnclaveAPIContainerHostMachineInfo do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :ip_on_host_machine, 4, type: :string, json_name: "ipOnHostMachine"
  field :grpc_port_on_host_machine, 5, type: :uint32, json_name: "grpcPortOnHostMachine"
end

defmodule EngineApi.EnclaveInfo do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_uuid, 1, type: :string, json_name: "enclaveUuid"
  field :name, 2, type: :string
  field :shortened_uuid, 3, type: :string, json_name: "shortenedUuid"

  field :containers_status, 4,
    type: EngineApi.EnclaveContainersStatus,
    json_name: "containersStatus",
    enum: true

  field :api_container_status, 5,
    type: EngineApi.EnclaveAPIContainerStatus,
    json_name: "apiContainerStatus",
    enum: true

  field :api_container_info, 6,
    type: EngineApi.EnclaveAPIContainerInfo,
    json_name: "apiContainerInfo"

  field :api_container_host_machine_info, 7,
    type: EngineApi.EnclaveAPIContainerHostMachineInfo,
    json_name: "apiContainerHostMachineInfo"

  field :creation_time, 8, type: Google.Protobuf.Timestamp, json_name: "creationTime"
  field :mode, 9, type: EngineApi.EnclaveMode, enum: true
end

defmodule EngineApi.GetEnclavesResponse.EnclaveInfoEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: EngineApi.EnclaveInfo
end

defmodule EngineApi.GetEnclavesResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_info, 1,
    repeated: true,
    type: EngineApi.GetEnclavesResponse.EnclaveInfoEntry,
    json_name: "enclaveInfo",
    map: true
end

defmodule EngineApi.EnclaveIdentifiers do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_uuid, 1, type: :string, json_name: "enclaveUuid"
  field :name, 2, type: :string
  field :shortened_uuid, 3, type: :string, json_name: "shortenedUuid"
end

defmodule EngineApi.GetExistingAndHistoricalEnclaveIdentifiersResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :allIdentifiers, 1, repeated: true, type: EngineApi.EnclaveIdentifiers
end

defmodule EngineApi.StopEnclaveArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_identifier, 1, type: :string, json_name: "enclaveIdentifier"
end

defmodule EngineApi.DestroyEnclaveArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_identifier, 1, type: :string, json_name: "enclaveIdentifier"
end

defmodule EngineApi.CleanArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :should_clean_all, 1, type: :bool, json_name: "shouldCleanAll"
end

defmodule EngineApi.EnclaveNameAndUuid do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :name, 1, type: :string
  field :uuid, 2, type: :string
end

defmodule EngineApi.CleanResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :removed_enclave_name_and_uuids, 1,
    repeated: true,
    type: EngineApi.EnclaveNameAndUuid,
    json_name: "removedEnclaveNameAndUuids"
end

defmodule EngineApi.GetServiceLogsArgs.ServiceUuidSetEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: :bool
end

defmodule EngineApi.GetServiceLogsArgs do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :enclave_identifier, 1, type: :string, json_name: "enclaveIdentifier"

  field :service_uuid_set, 2,
    repeated: true,
    type: EngineApi.GetServiceLogsArgs.ServiceUuidSetEntry,
    json_name: "serviceUuidSet",
    map: true

  field :follow_logs, 3, type: :bool, json_name: "followLogs"

  field :conjunctive_filters, 4,
    repeated: true,
    type: EngineApi.LogLineFilter,
    json_name: "conjunctiveFilters"

  field :return_all_logs, 5, type: :bool, json_name: "returnAllLogs"
  field :num_log_lines, 6, type: :uint32, json_name: "numLogLines"
end

defmodule EngineApi.GetServiceLogsResponse.ServiceLogsByServiceUuidEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: EngineApi.LogLine
end

defmodule EngineApi.GetServiceLogsResponse.NotFoundServiceUuidSetEntry do
  @moduledoc false

  use Protobuf, map: true, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :key, 1, type: :string
  field :value, 2, type: :bool
end

defmodule EngineApi.GetServiceLogsResponse do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :service_logs_by_service_uuid, 1,
    repeated: true,
    type: EngineApi.GetServiceLogsResponse.ServiceLogsByServiceUuidEntry,
    json_name: "serviceLogsByServiceUuid",
    map: true

  field :not_found_service_uuid_set, 2,
    repeated: true,
    type: EngineApi.GetServiceLogsResponse.NotFoundServiceUuidSetEntry,
    json_name: "notFoundServiceUuidSet",
    map: true
end

defmodule EngineApi.LogLine do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :line, 1, repeated: true, type: :string
end

defmodule EngineApi.LogLineFilter do
  @moduledoc false

  use Protobuf, protoc_gen_elixir_version: "0.12.0", syntax: :proto3

  field :operator, 1, type: EngineApi.LogLineOperator, enum: true
  field :text_pattern, 2, type: :string, json_name: "textPattern"
end

defmodule EngineApi.EngineService.Service do
  @moduledoc false

  use GRPC.Service, name: "engine_api.EngineService", protoc_gen_elixir_version: "0.12.0"

  rpc :GetEngineInfo, Google.Protobuf.Empty, EngineApi.GetEngineInfoResponse

  rpc :CreateEnclave, EngineApi.CreateEnclaveArgs, EngineApi.CreateEnclaveResponse

  rpc :GetEnclaves, Google.Protobuf.Empty, EngineApi.GetEnclavesResponse

  rpc :GetExistingAndHistoricalEnclaveIdentifiers,
      Google.Protobuf.Empty,
      EngineApi.GetExistingAndHistoricalEnclaveIdentifiersResponse

  rpc :StopEnclave, EngineApi.StopEnclaveArgs, Google.Protobuf.Empty

  rpc :DestroyEnclave, EngineApi.DestroyEnclaveArgs, Google.Protobuf.Empty

  rpc :Clean, EngineApi.CleanArgs, EngineApi.CleanResponse

  rpc :GetServiceLogs, EngineApi.GetServiceLogsArgs, stream(EngineApi.GetServiceLogsResponse)
end

defmodule EngineApi.EngineService.Stub do
  @moduledoc false

  use GRPC.Stub, service: EngineApi.EngineService.Service
end
defmodule Backend.Engine.Enclave do
  defstruct enclave_uuid: nil,
            name: nil,
            containers_status: nil,
            api_container_status: nil,
            api_container_host_machine_info: nil,
            creation_time: nil

  def from_api(%EngineApi.EnclaveInfo{
        enclave_uuid: uuid,
        name: name,
        containers_status: containers_status,
        api_container_status: apic_status,
        api_container_host_machine_info: %EngineApi.EnclaveAPIContainerHostMachineInfo{
          ip_on_host_machine: api_ip,
          grpc_port_on_host_machine: api_grpc_port
          },
        creation_time: timestamp
      }) do
    %Backend.Engine.Enclave{
      enclave_uuid: uuid,
      name: name,
      containers_status: String.replace(to_string(containers_status), "EnclaveContainersStatus_", ""),
      api_container_status: String.replace(to_string(apic_status), "EnclaveAPIContainerStatus_", ""),
      api_container_host_machine_info: {api_ip, api_grpc_port},
      creation_time:
        DateTime.from_unix!(timestamp.seconds * 1_000_000_000 + timestamp.nanos, :nanosecond)
    }
  end

  # TODOs: add batter error handlers, config engine host/port, reuse GRP channel (maybe a genserver)

  def list_enclaves() do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    channel
    |> EngineApi.EngineService.Stub.get_enclaves(%Google.Protobuf.Empty{})
    |> (fn {:ok, reply} -> reply end).()
    |> (fn %EngineApi.GetEnclavesResponse{enclave_info: encs} -> encs end).()
    |> Map.values()
    |> Enum.map(&from_api(&1))
  end

  def stop_enclave(enclave_id) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    # TODO: how to pattern match this? %Google.Protobuf.Empty{} = channel
    channel
    |> EngineApi.EngineService.Stub.stop_enclave(%EngineApi.StopEnclaveArgs{
      enclave_identifier: enclave_id
    })
  end

  def destroy_enclave(enclave_id) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    channel
    |> EngineApi.EngineService.Stub.destroy_enclave(%EngineApi.DestroyEnclaveArgs{
      enclave_identifier: enclave_id
    })
  end
end

defmodule Backend.Engine.Service do
  def list_services(%Backend.Engine.Enclave{api_container_host_machine_info: {host_ip, port}}) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("#{host_ip}:#{port}")

    req = %ApiContainerApi.GetServicesArgs{service_identifiers: %{}}

    channel
    |> ApiContainerApi.ApiContainerService.Stub.get_services(req)
    |> (fn {:ok, reply} -> reply end).()
    |> (fn %ApiContainerApi.GetServicesResponse{service_info: encs} -> encs end).()
    |> Map.values()
  end


  def get_log_stream(enclave_id, services, follow) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    # services = %EngineApi.GetServiceLogsArgs.ServiceUuidSetEntry{key: "4aa37eb8b57a4f3e870366be64090fc9", value: false}
    # services = %{"81d8f204ee474521a8c0f0829868be36" => false, "63005b464b0140b9a8fffa89581e7e9f" => false}
    services = Enum.reduce(services, %{}, fn service, map -> Map.put(map, service, true) end)
    req = %EngineApi.GetServiceLogsArgs{
      enclave_identifier: enclave_id,
      service_uuid_set: services,
      num_log_lines: 5,
      follow_logs: follow
    }
    # IO.inspect(Protobuf.encode(req))

    channel
      |> EngineApi.EngineService.Stub.get_service_logs(req)
      |> (fn {:ok, reply} -> reply end).()
      |> Stream.flat_map(fn({:ok, reply}) -> Map.values(reply.service_logs_by_service_uuid) end)
      |> Stream.map(&(&1.line))
  end
end

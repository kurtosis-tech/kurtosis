defmodule Backend.Engine.Service do
  def list_services() do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:55611")

    req = %ApiContainerApi.GetServicesArgs{service_identifiers: %{}}

    channel
    |> ApiContainerApi.ApiContainerService.Stub.get_services(req)
    |> (fn {:ok, reply} -> reply end).()
    |> (fn %ApiContainerApi.GetServicesResponse{service_info: encs} -> encs end).()
    |> Map.values()
  end


  def get_log_stream(enclave_id) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    services = %EngineApi.GetServiceLogsArgs.ServiceUuidSetEntry{}
    req = %EngineApi.GetServiceLogsArgs{enclave_identifier: enclave_id, service_uuid_set: %{}}

    foo = channel
    |> EngineApi.EngineService.Stub.get_service_logs(req)
    |> (fn {:ok, reply} -> reply end).()
    # |> (fn %ApiContainerApi.GetServicesResponse{service_info: encs} -> encs end).()
    foo.("", fn x -> {x, ""} end)
  end
end

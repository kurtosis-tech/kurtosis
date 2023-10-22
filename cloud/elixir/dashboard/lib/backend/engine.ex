defmodule Backend.Engine do


  # TODOs: add batter error handlers, config engine host/port, reuse GRP channel (maybe a genserver)

  def list_enclaves() do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    channel
      |> EngineApi.EngineService.Stub.get_enclaves(%Google.Protobuf.Empty{})
      |> (fn {:ok, reply} -> reply end).()
      |> (fn %EngineApi.GetEnclavesResponse{enclave_info: encs} -> encs end).()
      |> Map.values()
  end

  def stop_enclave(enclave_id) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    # TODO: how to pattern match this? %Google.Protobuf.Empty{} = channel
    channel
      |> EngineApi.EngineService.Stub.stop_enclave(%EngineApi.StopEnclaveArgs{enclave_identifier: enclave_id})
  end

  def destroy_enclave(enclave_id) do
    # Add a second arg to log all communication "interceptors: [GRPC.Client.Interceptors.Logger]"
    {:ok, channel} = GRPC.Stub.connect("localhost:9710")

    channel
      |> EngineApi.EngineService.Stub.destroy_enclave(%EngineApi.DestroyEnclaveArgs{enclave_identifier: enclave_id})
  end
end

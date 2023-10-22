defmodule Backend.Engine do
  def list_enclaves() do
    {:ok, channel} =
      GRPC.Stub.connect("localhost:9710", interceptors: [GRPC.Client.Interceptors.Logger])

    {:ok, reply} =
      channel
      |> EngineApi.EngineService.Stub.get_enclaves(%Google.Protobuf.Empty{})
    # pass tuple `timeout: :infinity` as a second arg to stay in IEx debugging
    %EngineApi.GetEnclavesResponse{enclave_info: encs} = reply
    Map.values(encs)
  end
end

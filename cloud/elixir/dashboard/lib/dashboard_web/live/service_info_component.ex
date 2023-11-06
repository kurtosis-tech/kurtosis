defmodule DashboardWeb.ServiceInfoComponent do
  use DashboardWeb, :live_component

  @impl true
  def mount(socket) do
    {:ok,
     socket
     |> stream_configure(:logs, dom_id: &"msgs-#{&1.id}")
     |> stream(:logs, [], at: -1, limit: -10)}
  end

  def handle_event("stop_enclave", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("destory_enclave", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("show_logs", %{"id" => service_uuid}, socket) do
    pid = self()
    Task.async(fn ->
      logs =
        Backend.Engine.Service.get_log_stream(
          socket.assigns.enclave.enclave_uuid,
          [socket.assigns.service.service_uuid],
          true
        )

      logs
      |> Stream.with_index()
      |> Stream.each(fn {x, id} ->
        send_update(pid, DashboardWeb.ServiceInfoComponent,
          id: "service_info",
          push_log: %{:id => id, :line => x}
        )
      end)
      |> Enum.to_list() # This forces the eval of the lazy stream

    end)

    {:noreply, socket}
  end

  @impl true
  def update(%{push_log: log}, socket) do
    {:ok, socket |> stream_insert(:logs, log)}
  end

  @impl true
  def update(regular_assigns, socket) do
    {:ok, assign(socket, regular_assigns)}
  end
end

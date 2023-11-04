defmodule DashboardWeb.DashboardWeb.EngineApp do
  use DashboardWeb, :live_view

  def mount(_params, _session, socket) do
    enclaves = Backend.Engine.Enclave.list_enclaves()
    send(self(), :loop)
    {:ok, assign(socket, %{enclaves: enclaves, services: [], parent: self()})}
  end

  def handle_event(ev, _params, socket) do
    IO.inspect("Got event #{ev}")
    {:noreply, socket}
  end

  def handle_info(:refresh_enclaves, socket) do
    enclaves = Backend.Engine.Enclave.list_enclaves()
    {:noreply, assign(socket, %{enclaves: enclaves})}
  end

  def handle_info({:inspect_services, services}, socket) do
    IO.inspect(services)
    {:noreply, assign(socket, %{services: services})}
  end

  def handle_info(:loop, socket) do
    send(self(), :refresh_enclaves)
    Process.send_after(self(), :loop, 1000)
    {:noreply, socket}
  end
end

defmodule DashboardWeb.DashboardWeb.EngineApp do
  use DashboardWeb, :live_view

  def mount(_params, _session, socket) do
    enclaves = Backend.Engine.Enclave.list_enclaves()
    send(self(), :loop)
    {:ok, assign(socket, %{enclaves: enclaves, services: nil, service_info: nil, select_enclave: nil, parent: self()})}
  end

  def handle_info(:refresh_enclaves, socket) do
    enclaves = Backend.Engine.Enclave.list_enclaves()
    {:noreply, assign(socket, %{enclaves: enclaves, services: nil, service_info: nil, select_enclave: nil})}
  end

  def handle_info({:inspect_services, {select_enclave, services}}, socket) do
    {:noreply, assign(socket, %{enclaves: nil, services: services, service_info: nil, select_enclave: select_enclave})}
  end

  def handle_info({:service_info, service_info}, socket) do
    {:noreply, assign(socket, %{enclaves: nil, services: nil, service_info: service_info})}
  end

  def handle_info(:loop, socket) do
    if socket.assigns.enclaves do send(self(), :refresh_enclaves) end
    Process.send_after(self(), :loop, 1000)
    {:noreply, socket}
  end

  def handle_info({:inspect_services_2, enclave_uuid}, socket) do
    select_enclave =
      Backend.Engine.Enclave.list_enclaves()
      |> Enum.find(fn x ->
        if x.enclave_uuid == enclave_uuid do
          x
        end
      end)

    services = Backend.Engine.Service.list_services(select_enclave)
    {:noreply, assign(socket, %{enclaves: nil, services: services, service_info: nil, select_enclave: select_enclave})}
  end

  def handle_event("list_enclaves", _, socket) do
    send(self(), :refresh_enclaves)
    {:noreply, socket}
  end

  def handle_event("list_services", %{"enclave_uuid" => enclave_uuid}, socket) do
    send(self(), {:inspect_services_2, enclave_uuid})
    {:noreply, socket}
  end

  def handle_event(ev, _params, socket) do
    IO.inspect("Got event #{ev}")
    {:noreply, socket}
  end
end

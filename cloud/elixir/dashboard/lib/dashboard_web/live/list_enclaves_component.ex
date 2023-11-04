defmodule DashboardWeb.ListEnclavesComponent do
  use DashboardWeb, :live_component

  def handle_event("stop_enclave", %{"id" => enclave_uuid}, socket) do
    Backend.Engine.Enclave.stop_enclave(enclave_uuid)
    send(socket.assigns.parent, :refresh_enclaves)
    {:noreply, socket}
  end

  def handle_event("destory_enclave", %{"id" => enclave_uuid}, socket) do
    Backend.Engine.Enclave.destroy_enclave(enclave_uuid)
    send(socket.assigns.parent, :refresh_enclaves)
    {:noreply, socket}
  end

  def handle_event("inspect_enclave", %{"id" => enclave_uuid}, socket) do
    enclave = socket.assigns.items |> Enum.find(fn x -> if x.enclave_uuid == enclave_uuid do; x; end; end)
    services = Backend.Engine.Service.list_services(enclave)
    IO.inspect(services)
    send(socket.assigns.parent, {:inspect_services, services})
    {:noreply, socket}
  end

end

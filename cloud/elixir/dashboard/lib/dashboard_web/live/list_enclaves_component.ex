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

end

defmodule DashboardWeb.ListEnclavesComponent do
  use DashboardWeb, :live_component

  def handle_event("stop_enclave", %{"id" => enclave_uuid}, socket) do
    Backend.Engine.stop_enclave(enclave_uuid)
    {:noreply, socket}
  end

  def handle_event("destory_enclave", %{"id" => enclave_uuid}, socket) do
    Backend.Engine.destroy_enclave(enclave_uuid)
    {:noreply, socket}
  end

end

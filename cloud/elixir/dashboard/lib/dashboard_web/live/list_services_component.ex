defmodule DashboardWeb.ListServicesComponent do
  use DashboardWeb, :live_component

  def handle_event("stop_enclave", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("destory_enclave", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("inspect_enclave", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

end

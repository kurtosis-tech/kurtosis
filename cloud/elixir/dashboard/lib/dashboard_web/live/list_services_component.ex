defmodule DashboardWeb.ListServicesComponent do
  use DashboardWeb, :live_component

  def handle_event("stop_service", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("destory_service", %{"id" => service_uuid}, socket) do
    {:noreply, socket}
  end

  def handle_event("inspect_service", %{"id" => service_uuid}, socket) do
    {:noreply, socket}

    service =
      socket.assigns.items
      |> Enum.find(fn x ->
        if x.service_uuid == service_uuid do
          x
        end
      end)

    # services = Backend.Engine.Service.list_services(enclave)
    IO.inspect(service)
    send(socket.assigns.parent, {:service_info, service})
    {:noreply, socket}
  end
end

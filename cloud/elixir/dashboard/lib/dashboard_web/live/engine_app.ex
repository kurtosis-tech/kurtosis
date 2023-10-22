defmodule DashboardWeb.DashboardWeb.EngineApp do
  use DashboardWeb, :live_view

  def mount(_params, _session, socket) do
    enclaves = Backend.Engine.list_enclaves()
    {:ok, assign(socket, :enclaves, enclaves)}
  end

  def handle_event(ev, _params, socket) do
    IO.inspect("Got event #{ev}")
    {:noreply, socket}
  end
end

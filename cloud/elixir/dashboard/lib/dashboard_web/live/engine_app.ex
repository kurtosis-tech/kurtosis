defmodule DashboardWeb.DashboardWeb.EngineApp do
  alias DashboardWeb.DashboardWeb.EngineApp
  use DashboardWeb, :live_view

  def mount(_params, _session, socket) do
    enclaves = Backend.Engine.list_enclaves()
    {:ok, assign(socket, :enclaves, enclaves)}
  end

  def handle_event("inc_temperature", _params, socket) do
    {:noreply, update(socket, :temperature, &(&1 + 1))}
  end
end

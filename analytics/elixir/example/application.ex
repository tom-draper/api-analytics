defmodule MyApp.Application do
  use Application

  @impl true
  def start(_type, _args) do
    children = [
      # Start ApiAnalytics before the endpoint so it's ready to receive requests
      {ApiAnalytics, api_key: System.get_env("API_ANALYTICS_KEY")},
      MyAppWeb.Endpoint
    ]

    opts = [strategy: :one_for_one, name: MyApp.Supervisor]
    Supervisor.start_link(children, opts)
  end
end

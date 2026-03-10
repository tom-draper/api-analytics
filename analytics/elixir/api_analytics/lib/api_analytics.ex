defmodule ApiAnalytics do
  @moduledoc """
  API Analytics for Phoenix and Plug applications.

  Add `ApiAnalytics` to your supervision tree (typically in `application.ex`):

      children = [
        {ApiAnalytics, api_key: System.get_env("API_ANALYTICS_KEY")},
        MyAppWeb.Endpoint
      ]

  Then add `ApiAnalytics.Plug` to your endpoint or router pipeline:

      # In endpoint.ex (logs all requests)
      plug ApiAnalytics.Plug

      # Or in router.ex (logs only API pipeline)
      pipeline :api do
        plug :accepts, ["json"]
        plug ApiAnalytics.Plug
      end

  ## Options

    * `:api_key` - Your API Analytics key (required)
    * `:privacy_level` - Controls IP address handling (default: `0`)
      * `0` - IP stored and used for location
      * `1` - IP used for location only, then discarded
      * `2` - IP never sent to the server
    * `:server_url` - Override for self-hosting (default: `"https://www.apianalytics-server.com/"`)
  """

  use GenServer

  @server_url "https://www.apianalytics-server.com/"
  @flush_interval 60_000

  @type t :: %__MODULE__{
          api_key: String.t(),
          privacy_level: non_neg_integer(),
          server_url: String.t(),
          get_path: (Plug.Conn.t() -> String.t()) | nil,
          get_hostname: (Plug.Conn.t() -> String.t()) | nil,
          get_ip_address: (Plug.Conn.t() -> String.t() | nil) | nil,
          get_user_agent: (Plug.Conn.t() -> String.t()) | nil,
          get_user_id: (Plug.Conn.t() -> String.t() | nil) | nil
        }

  defstruct api_key: "",
            privacy_level: 0,
            server_url: @server_url,
            get_path: nil,
            get_hostname: nil,
            get_ip_address: nil,
            get_user_agent: nil,
            get_user_id: nil

  @doc """
  Starts the ApiAnalytics GenServer and links it to the calling process.
  """
  def start_link(opts) do
    {name, opts} = Keyword.pop(opts, :name, __MODULE__)
    GenServer.start_link(__MODULE__, opts, name: name)
  end

  @doc """
  Logs a request. Called automatically by `ApiAnalytics.Plug`.
  """
  def log(request_data, server \\ __MODULE__) do
    GenServer.cast(server, {:log, request_data})
  end

  @impl GenServer
  def init(opts) do
    config = struct(__MODULE__, opts)
    schedule_flush()
    {:ok, %{config: config, requests: []}}
  end

  @impl GenServer
  def handle_cast({:log, request_data}, state) do
    {:noreply, %{state | requests: [request_data | state.requests]}}
  end

  @impl GenServer
  def handle_info(:flush, state) do
    if state.requests != [] do
      flush_requests(state.config, state.requests)
    end

    schedule_flush()
    {:noreply, %{state | requests: []}}
  end

  defp schedule_flush do
    Process.send_after(self(), :flush, @flush_interval)
  end

  defp flush_requests(config, requests) do
    payload =
      Jason.encode!(%{
        api_key: config.api_key,
        requests: Enum.reverse(requests),
        framework: "Phoenix",
        privacy_level: config.privacy_level
      })

    url = String.trim_trailing(config.server_url, "/") <> "/api/log-request"

    :httpc.request(
      :post,
      {String.to_charlist(url), [], ~c"application/json", payload},
      [{:timeout, 10_000}],
      []
    )
  end
end

defmodule ApiAnalytics.Plug do
  @moduledoc """
  Plug middleware that logs each request to `ApiAnalytics`.
  """

  @behaviour Plug

  alias Plug.Conn

  @impl Plug
  def init(opts), do: Keyword.get(opts, :server, ApiAnalytics)

  @impl Plug
  def call(conn, server) do
    start = System.monotonic_time(:millisecond)

    Conn.register_before_send(conn, fn conn ->
      elapsed = System.monotonic_time(:millisecond) - start

      request_data = %{
        hostname: conn.host,
        ip_address: get_ip_address(conn),
        path: conn.request_path,
        user_agent: get_header(conn, "user-agent"),
        method: conn.method,
        status: conn.status,
        response_time: elapsed,
        user_id: nil,
        created_at: DateTime.utc_now() |> DateTime.to_iso8601()
      }

      ApiAnalytics.log(request_data, server)
      conn
    end)
  end

  defp get_ip_address(conn) do
    case Conn.get_req_header(conn, "cf-connecting-ip") do
      [ip | _] ->
        ip

      [] ->
        case Conn.get_req_header(conn, "x-forwarded-for") do
          [forwarded | _] ->
            forwarded |> String.split(",") |> List.first() |> String.trim()

          [] ->
            case Conn.get_req_header(conn, "x-real-ip") do
              [ip | _] -> ip
              [] -> conn.remote_ip |> :inet.ntoa() |> List.to_string()
            end
        end
    end
  end

  defp get_header(conn, header) do
    case Conn.get_req_header(conn, header) do
      [value | _] -> value
      [] -> ""
    end
  end
end

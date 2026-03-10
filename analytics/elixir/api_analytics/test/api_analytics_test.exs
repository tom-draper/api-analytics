defmodule ApiAnalyticsTest do
  use ExUnit.Case, async: false
  use Plug.Test

  @opts ApiAnalytics.Plug.init(server: ApiAnalyticsTest.Server)

  setup do
    {:ok, pid} =
      start_supervised(
        {ApiAnalytics,
         name: ApiAnalyticsTest.Server,
         api_key: "test-api-key",
         server_url: "http://localhost:99999/"}
      )

    {:ok, server: pid}
  end

  defp call(method, path, status \\ 200) do
    conn =
      conn(method, path)
      |> ApiAnalytics.Plug.call(@opts)

    %{conn | status: status}
    |> send_resp(status, "")
  end

  test "passes through GET request" do
    conn = call(:get, "/api/users")
    assert conn.status == 200
  end

  test "passes through POST request" do
    conn = call(:post, "/api/users", 201)
    assert conn.status == 201
  end

  test "passes through 404" do
    conn = call(:get, "/not-found", 404)
    assert conn.status == 404
  end

  test "passes through 500" do
    conn = call(:get, "/api/error", 500)
    assert conn.status == 500
  end

  test "buffers requests" do
    call(:get, "/api/users")
    call(:get, "/api/users")
    call(:post, "/api/users", 201)

    state = :sys.get_state(ApiAnalyticsTest.Server)
    assert length(state.requests) == 3
  end

  test "captures correct path" do
    call(:get, "/api/users/42")

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.path == "/api/users/42"
  end

  test "captures correct method" do
    call(:delete, "/api/users/1", 204)

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.method == "DELETE"
  end

  test "captures correct status" do
    call(:get, "/api/users", 200)

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.status == 200
  end

  test "captures x-forwarded-for header" do
    conn =
      conn(:get, "/api/users")
      |> put_req_header("x-forwarded-for", "1.2.3.4, 5.6.7.8")
      |> ApiAnalytics.Plug.call(@opts)

    %{conn | status: 200} |> send_resp(200, "")

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.ip_address == "1.2.3.4"
  end

  test "captures cf-connecting-ip header" do
    conn =
      conn(:get, "/api/users")
      |> put_req_header("cf-connecting-ip", "9.9.9.9")
      |> ApiAnalytics.Plug.call(@opts)

    %{conn | status: 200} |> send_resp(200, "")

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.ip_address == "9.9.9.9"
  end

  test "captures user agent" do
    conn =
      conn(:get, "/api/users")
      |> put_req_header("user-agent", "TestAgent/1.0")
      |> ApiAnalytics.Plug.call(@opts)

    %{conn | status: 200} |> send_resp(200, "")

    state = :sys.get_state(ApiAnalyticsTest.Server)
    [request | _] = state.requests
    assert request.user_agent == "TestAgent/1.0"
  end
end

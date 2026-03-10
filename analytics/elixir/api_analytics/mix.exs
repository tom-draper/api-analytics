defmodule ApiAnalytics.MixProject do
  use Mix.Project

  @version "1.0.0"
  @source_url "https://github.com/tom-draper/api-analytics"

  def project do
    [
      app: :api_analytics,
      version: @version,
      elixir: "~> 1.12",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      description: "Monitoring and analytics for Phoenix and Plug API applications.",
      package: package(),
      docs: docs()
    ]
  end

  def application do
    [
      extra_applications: [:logger, :inets, :ssl]
    ]
  end

  defp deps do
    [
      {:plug, "~> 1.0"},
      {:jason, "~> 1.0"},
      {:ex_doc, "~> 0.27", only: :dev, runtime: false},
      {:plug_cowboy, "~> 2.0", only: :test}
    ]
  end

  defp package do
    [
      name: "api_analytics",
      licenses: ["MIT"],
      links: %{"GitHub" => @source_url},
      maintainers: ["Tom Draper"]
    ]
  end

  defp docs do
    [
      main: "ApiAnalytics",
      source_url: @source_url
    ]
  end
end

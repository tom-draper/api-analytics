require_relative 'boot'
require 'rails/all'

Bundler.require(*Rails.groups)

module RailsExample
  class Application < Rails::Application
    config.load_defaults 7.1
    config.api_only = true

    config.middleware.use ::Analytics::Rails, ENV['API_KEY']  # Add middleware
  end
end

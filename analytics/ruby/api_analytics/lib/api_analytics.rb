# frozen_string_literal: true

require 'uri'
require 'net/http'
require 'json'

module Analytics
  class Middleware
    def initialize(app, api_key)
      @app = app
      @api_key = api_key
    end

    def call(env)
      start = Time.now
      status, headers, response = @app.call(env)

      data = {
        api_key: @api_key,
        hostname: env['HTTP_HOST'],
        path: env['REQUEST_PATH'],
        user_agent: env['HTTP_USER_AGENT'],
        method: env['REQUEST_METHOD'],
        status: status,
        framework: @framework,
        response_time: (Time.now - start).to_f.round,
      }

      Thread.new {
        log_request(data)
      }

      [status, headers, response]
    end

    private

    def log_request(data)
      uri = URI('https://api-analytics-server.vercel.app/api/log-request')
      res = Net::HTTP.post(uri, data.to_json)
    end
  end

  private_constant :Middleware
  
  class Rails < Middleware
    def initialize(app, api_key)
      super(app, api_key)
      @framework = "Rails"
    end
  end

  class Sinatra < Middleware
    def initialize(app, api_key)
      super(app, api_key)
      @framework = "Sinatra"
    end
  end
end




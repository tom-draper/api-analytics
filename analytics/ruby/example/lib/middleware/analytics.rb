require 'net/http'
require 'json'

module Middleware
  class Analytics
    def initialize(app)
      @app = app
    end

    def call(env)
      start = Time.now
      status, headers, response = @app.call(env)

      json = {
        api_key: "9e8820b1-6308-40f8-8423-d913c203b92d",
        hostname: env['HTTP_HOST'],
        path: env['REQUEST_PATH'],
        user_agent: env['HTTP_USER_AGENT'],
        method: env['REQUEST_METHOD'],
        status: status,
        framework: "Rails",
        response_time: (Time.now - start).to_f.round,
      }

      Thread.new {
        log_request(json)
      }

      [status, headers, response]
    end

    private

    def log_request(json)
      uri = URI('http://localhost:8080/api/log-request')
      http = Net::HTTP.new(uri.host, uri.port)
      Rails.logger.info("#{uri.host} #{uri.port} #{uri.path}")
      req = Net::HTTP::Post.new(uri.path, 'Content-Type' => 'application/json')
      req.body = json.to_json
      res = http.request(req)
    rescue => e
    end
  end
end


require 'net/http'
require 'json'

module Middleware
  class Analytics
    def initialize(app)
      @app = app
      @api_key = Rails.application.secrets.ANALYTICS_API_KEY
      raise StandardError.new 'ANALYTICS_API_KEY secret unset.' if @api_key.nil?
    end

    def call(env)
      start = Time.now
      status, headers, response = @app.call(env)

      json = {
        api_key: @api_key,
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
      uri = URI('https://api-analytics-server.vercel.app/api/log-request')
      http = Net::HTTP.new(uri.host, uri.port)
      # Rails.logger.info("#{uri.host} #{uri.port} #{uri.path}")
      req = Net::HTTP::Post.new(uri.path, 'Content-Type' => 'application/json')
      req.body = json.to_json
      res = http.request(req)
    rescue => e
    end
  end
end


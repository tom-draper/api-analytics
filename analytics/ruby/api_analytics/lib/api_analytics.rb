# frozen_string_literal: true

require 'uri'
require 'net/http'
require 'json'
require 'time'

module Analytics
  class Middleware
    def initialize(app, api_key, config = Config.new)
      @app = app
      @api_key = api_key
      @config = config
      @framework = "Middleware"
      @requests = []
      @last_posted = Time.now
      @mutex = Mutex.new
    end

    def call(env)
      start = Time.now
      status, headers, response = @app.call(env)

      request_data = {
        hostname: @config.get_hostname.call(env),
        ip_address: get_ip_address(env),
        path: @config.get_path.call(env),
        user_agent: @config.get_user_agent.call(env),
        method: env['REQUEST_METHOD'],
        status: status,
        response_time: ((Time.now - start) * 1000).round,
        user_id: @config.get_user_id.call(env),
        created_at: Time.now.utc.iso8601
      }

      log_request(request_data)

      [status, headers, response]
    end

    private

    def get_ip_address(env)
      return nil if @config.privacy_level >= 2

      @config.get_ip_address.call(env)
    end

    def post_requests(requests)
      return if @api_key.to_s.empty?

      payload = {
        api_key: @api_key,
        requests: requests,
        framework: @framework,
        privacy_level: @config.privacy_level
      }

      url = @config.server_url.end_with?('/') ? @config.server_url + 'api/log-request' : @config.server_url + '/api/log-request'
      uri = URI(url)

      Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == 'https', open_timeout: 10, read_timeout: 10) do |http|
        request = Net::HTTP::Post.new(uri, 'Content-Type' => 'application/json')
        request.body = payload.to_json
        http.request(request)
      end
    end

    def log_request(request_data)
      now = Time.now
      requests_to_post = nil

      @mutex.synchronize do
        @requests.push(request_data)
        if (now - @last_posted) > 60.0
          requests_to_post = @requests.dup
          @requests = []
          @last_posted = now
        end
      end

      Thread.new { post_requests(requests_to_post) } if requests_to_post
    end
  end

  Config = Struct.new(:privacy_level, :server_url, :get_path, :get_ip_address, :get_hostname, :get_user_agent, :get_user_id) do
    def initialize(
      privacy_level = 0,
      server_url = 'https://www.apianalytics-server.com',
      get_path = ->(env) { env['REQUEST_PATH'] },
      get_ip_address = lambda { |env|
        env['HTTP_CF_CONNECTING_IP'] ||
          (env['HTTP_X_FORWARDED_FOR'] && env['HTTP_X_FORWARDED_FOR'].split(',').first&.strip) ||
          env['HTTP_X_REAL_IP'] ||
          env['REMOTE_ADDR']
      },
      get_hostname = ->(env) { env['HTTP_HOST'] },
      get_user_agent = ->(env) { env['HTTP_USER_AGENT'] },
      get_user_id = ->(_env) { nil }
    )
      self.privacy_level = privacy_level
      self.server_url = server_url
      self.get_path = get_path
      self.get_ip_address = get_ip_address
      self.get_hostname = get_hostname
      self.get_user_agent = get_user_agent
      self.get_user_id = get_user_id
    end
  end

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

# frozen_string_literal: true

require 'uri'
require 'net/http'
require 'json'
require 'time'

module Analytics
  class Middleware
    FLUSH_INTERVAL = 60

    def initialize(app, api_key, config = Config.new)
      @app = app
      @api_key = api_key
      @config = config
      @framework = "Middleware"
      @requests = []
      @last_posted = Time.now
      @mutex = Mutex.new
      @flusher_started = false
    end

    def call(env)
      start = Time.now
      status, headers, response = @app.call(env)

      # Collecting analytics must never break the host response, so any failure
      # in a data mapper or the buffer is swallowed here.
      begin
        ensure_flusher

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
      rescue StandardError
        # Best-effort: drop this request's analytics rather than raise.
      end

      [status, headers, response]
    end

    # flush posts any buffered requests immediately. Exposed so an application
    # can drain the buffer from its own shutdown hook and avoid losing the final
    # batch on a graceful restart.
    def flush
      requests_to_post = nil
      @mutex.synchronize do
        unless @requests.empty?
          requests_to_post = @requests
          @requests = []
          @last_posted = Time.now
        end
      end
      post_requests(requests_to_post) if requests_to_post
    end

    private

    # ensure_flusher lazily starts a background thread that flushes buffered
    # requests every FLUSH_INTERVAL seconds, so a partial batch is not held
    # indefinitely when traffic goes idle. It is started on the first request
    # (rather than at construction) so the thread lives in the worker process
    # after a forking server (e.g. Puma with preload) spawns it.
    def ensure_flusher
      return if @flusher_started

      @mutex.synchronize do
        return if @flusher_started

        @flusher_started = true
      end

      Thread.new do
        loop do
          sleep(FLUSH_INTERVAL)
          begin
            flush
          rescue StandardError
            # Best-effort background flush.
          end
        end
      end
    end

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
    def initialize(app, api_key, config = Config.new)
      super(app, api_key, config)
      @framework = "Rails"
    end
  end

  class Sinatra < Middleware
    def initialize(app, api_key, config = Config.new)
      super(app, api_key, config)
      @framework = "Sinatra"
    end
  end
end

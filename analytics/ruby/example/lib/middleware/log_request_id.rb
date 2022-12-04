require 'securerandom'

module Middleware
  class LogRequestId
    REQUEST_ID = "X-Request-Id"

    def initialize(app)
      @app = app
    end

    def call(env)
      request_id = generate_request_id(env)
      @app.call(env).tap do |_status, headers, _body| 
        headers[@header] = request_id
      end
      @app.call(env)
    end

    private

    def generate_request_id(env)
      req = ActionDispatch::Request.new env
      request_id = req.headers[REQUEST_ID]
      if request_id.presence
        request_id.gsub(/[^\w\-@]/, "").first(255)
      else
        build_request_id
      end
    end

    def build_request_id
      SecureRandom.uuid
    end
  end
end

module Middleware
  class MeasureLogs
    def initialize(app)
      @app = app
    end

    def call(env)
      start_time = Time.now
      status, headers, response = @app.call(env)
      end_time = Time.now
      total_time = calculate_total_time(start_time, end_time)
      # log_status(total_time, status, env)

      [status, headers, response]
    end

    private

    # def log_status(total_time, status, env)
    #   controller_name = env['action_controller.instance'].class.name
    #   action = env['action_controller.instance'].action_name
    #   Rails.logger.info("----- [HTTP Request] status=#{status} time=#{total_time.to_s} processed_by=#{controller_name}\##{action} -----")
    # end

    def calculate_total_time(start_time, end_time)
      end_time - start_time
    end
  end
end

# frozen_string_literal: true

require 'spec_helper'
require 'rack/mock'

RSpec.describe Analytics do
  let(:ok_app) { ->(_env) { [200, { 'Content-Type' => 'application/json' }, ['{"message":"Hello"}']] } }
  let(:api_key) { 'test-api-key' }

  def mock_env(overrides = {})
    Rack::MockRequest.env_for('/', {
      'REQUEST_METHOD' => 'GET',
      'REQUEST_PATH' => '/',
      'HTTP_HOST' => 'example.com',
      'HTTP_USER_AGENT' => 'TestAgent/1.0',
      'REMOTE_ADDR' => '1.2.3.4'
    }.merge(overrides))
  end

  describe Analytics::Config do
    subject(:config) { Analytics::Config.new }

    it 'defaults privacy_level to 0' do
      expect(config.privacy_level).to eq(0)
    end

    it 'defaults server_url to the analytics server' do
      expect(config.server_url).to eq('https://www.apianalytics-server.com')
    end

    it 'defaults user_id to nil' do
      expect(config.get_user_id.call(mock_env)).to be_nil
    end

    it 'extracts path from env' do
      env = mock_env('REQUEST_PATH' => '/users')
      expect(config.get_path.call(env)).to eq('/users')
    end

    it 'extracts hostname from env' do
      env = mock_env('HTTP_HOST' => 'example.com')
      expect(config.get_hostname.call(env)).to eq('example.com')
    end

    it 'extracts user agent from env' do
      env = mock_env('HTTP_USER_AGENT' => 'Mozilla/5.0')
      expect(config.get_user_agent.call(env)).to eq('Mozilla/5.0')
    end

    describe 'IP address extraction' do
      it 'falls back to REMOTE_ADDR when no proxy headers present' do
        env = mock_env('REMOTE_ADDR' => '1.2.3.4')
        expect(config.get_ip_address.call(env)).to eq('1.2.3.4')
      end

      it 'prefers CF-Connecting-IP over all other headers' do
        env = mock_env(
          'HTTP_CF_CONNECTING_IP' => '5.6.7.8',
          'HTTP_X_FORWARDED_FOR' => '9.10.11.12',
          'HTTP_X_REAL_IP' => '13.14.15.16',
          'REMOTE_ADDR' => '1.2.3.4'
        )
        expect(config.get_ip_address.call(env)).to eq('5.6.7.8')
      end

      it 'prefers X-Forwarded-For over X-Real-IP and REMOTE_ADDR' do
        env = mock_env(
          'HTTP_X_FORWARDED_FOR' => '9.10.11.12, 13.14.15.16',
          'HTTP_X_REAL_IP' => '5.6.7.8',
          'REMOTE_ADDR' => '1.2.3.4'
        )
        expect(config.get_ip_address.call(env)).to eq('9.10.11.12')
      end

      it 'uses first IP from X-Forwarded-For when multiple are present' do
        env = mock_env('HTTP_X_FORWARDED_FOR' => '9.10.11.12, 13.14.15.16, 17.18.19.20')
        expect(config.get_ip_address.call(env)).to eq('9.10.11.12')
      end

      it 'prefers X-Real-IP over REMOTE_ADDR' do
        env = mock_env(
          'HTTP_X_REAL_IP' => '5.6.7.8',
          'REMOTE_ADDR' => '1.2.3.4'
        )
        expect(config.get_ip_address.call(env)).to eq('5.6.7.8')
      end
    end
  end

  describe Analytics::Middleware do
    subject(:middleware) { Analytics::Middleware.new(ok_app, api_key) }

    before { allow(middleware).to receive(:post_requests) }

    it 'passes through status unchanged' do
      status, = middleware.call(mock_env)
      expect(status).to eq(200)
    end

    it 'passes through headers unchanged' do
      _, headers, = middleware.call(mock_env)
      expect(headers['Content-Type']).to eq('application/json')
    end

    it 'passes through response body unchanged' do
      _, _, response = middleware.call(mock_env)
      expect(response).to eq(['{"message":"Hello"}'])
    end

    it 'passes through non-200 status codes' do
      app = ->(_env) { [404, {}, ['Not Found']] }
      m = Analytics::Middleware.new(app, api_key)
      allow(m).to receive(:post_requests)
      status, = m.call(mock_env)
      expect(status).to eq(404)
    end

    it 'logs request to buffer' do
      middleware.call(mock_env)
      expect(middleware.instance_variable_get(:@requests).length).to eq(1)
    end

    it 'records response time as a non-negative integer in milliseconds' do
      middleware.call(mock_env)
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:response_time]).to be_a(Integer)
      expect(data[:response_time]).to be >= 0
    end

    it 'records the correct HTTP method' do
      middleware.call(mock_env('REQUEST_METHOD' => 'POST'))
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:method]).to eq('POST')
    end

    it 'records the correct status code' do
      middleware.call(mock_env)
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:status]).to eq(200)
    end

    it 'records the path' do
      middleware.call(mock_env('REQUEST_PATH' => '/users'))
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:path]).to eq('/users')
    end

    it 'records the hostname' do
      middleware.call(mock_env('HTTP_HOST' => 'api.example.com'))
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:hostname]).to eq('api.example.com')
    end

    it 'records the user agent' do
      middleware.call(mock_env('HTTP_USER_AGENT' => 'Mozilla/5.0'))
      data = middleware.instance_variable_get(:@requests).first
      expect(data[:user_agent]).to eq('Mozilla/5.0')
    end

    it 'records created_at as an ISO8601 timestamp' do
      middleware.call(mock_env)
      data = middleware.instance_variable_get(:@requests).first
      expect { Time.iso8601(data[:created_at]) }.not_to raise_error
    end

    describe 'IP address and privacy' do
      it 'records ip_address at privacy level 0' do
        middleware.call(mock_env('REMOTE_ADDR' => '1.2.3.4'))
        data = middleware.instance_variable_get(:@requests).first
        expect(data[:ip_address]).to eq('1.2.3.4')
      end

      it 'records ip_address at privacy level 1' do
        config = Analytics::Config.new(1)
        m = Analytics::Middleware.new(ok_app, api_key, config)
        allow(m).to receive(:post_requests)
        m.call(mock_env('REMOTE_ADDR' => '1.2.3.4'))
        data = m.instance_variable_get(:@requests).first
        expect(data[:ip_address]).to eq('1.2.3.4')
      end

      it 'suppresses ip_address at privacy level 2' do
        config = Analytics::Config.new(2)
        m = Analytics::Middleware.new(ok_app, api_key, config)
        allow(m).to receive(:post_requests)
        m.call(mock_env('REMOTE_ADDR' => '1.2.3.4'))
        data = m.instance_variable_get(:@requests).first
        expect(data[:ip_address]).to be_nil
      end
    end

    describe 'custom mappers' do
      it 'uses a custom get_user_id mapper' do
        config = Analytics::Config.new(
          0,
          Analytics::Config.new.server_url,
          Analytics::Config.new.get_path,
          Analytics::Config.new.get_ip_address,
          Analytics::Config.new.get_hostname,
          Analytics::Config.new.get_user_agent,
          ->(_env) { 'custom-user-123' }
        )
        m = Analytics::Middleware.new(ok_app, api_key, config)
        allow(m).to receive(:post_requests)
        m.call(mock_env)
        data = m.instance_variable_get(:@requests).first
        expect(data[:user_id]).to eq('custom-user-123')
      end
    end

    describe 'request batching' do
      it 'does not flush before 60 seconds have elapsed' do
        expect(middleware).not_to receive(:post_requests)
        middleware.call(mock_env)
        expect(middleware.instance_variable_get(:@requests).length).to eq(1)
      end

      it 'flushes buffer after 60 seconds have elapsed' do
        middleware.instance_variable_set(:@last_posted, Time.now - 61)
        expect(middleware).to receive(:post_requests)
        middleware.call(mock_env)
        expect(middleware.instance_variable_get(:@requests)).to be_empty
      end

      it 'accumulates multiple requests before flushing' do
        3.times { middleware.call(mock_env) }
        expect(middleware.instance_variable_get(:@requests).length).to eq(3)
      end

      it 'does not flush when api_key is empty' do
        m = Analytics::Middleware.new(ok_app, '')
        expect(m).not_to receive(:post_requests)
        m.instance_variable_set(:@last_posted, Time.now - 61)
        m.call(mock_env)
      end
    end
  end

  describe Analytics::Rails do
    subject(:middleware) { Analytics::Rails.new(ok_app, api_key) }

    it 'sets framework to Rails' do
      expect(middleware.instance_variable_get(:@framework)).to eq('Rails')
    end

    it 'passes through requests correctly' do
      allow(middleware).to receive(:post_requests)
      status, = middleware.call(mock_env)
      expect(status).to eq(200)
    end
  end

  describe Analytics::Sinatra do
    subject(:middleware) { Analytics::Sinatra.new(ok_app, api_key) }

    it 'sets framework to Sinatra' do
      expect(middleware.instance_variable_get(:@framework)).to eq('Sinatra')
    end

    it 'passes through requests correctly' do
      allow(middleware).to receive(:post_requests)
      status, = middleware.call(mock_env)
      expect(status).to eq(200)
    end
  end
end

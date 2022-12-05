require 'sinatra'
require 'api_analytics'
require 'dotenv'

require './api_analytics.rb'

Dotenv.load()

use Analytics::Sinatra, ENV['API_KEY']

before do
    content_type 'application/json'
end

get '/' do
    {message: 'Hello World!'}.to_json
end
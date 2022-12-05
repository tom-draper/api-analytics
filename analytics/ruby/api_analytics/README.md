# API Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
gem install api_analytics
```

#### Rails

Add the analytics middleware to your rails application in `config/application.rb`.

```ruby
require 'rails'
require 'api_analytics'

Bundler.require(*Rails.groups)

module RailsMiddleware
  class Application < Rails::Application
    config.load_defaults 6.1
    config.api_only = true

    config.middleware.use ::Analytics::Rails, <api_key> # Add middleware
  end
end
```

#### Sinatra

```ruby
require 'sinatra'
require 'api_analytics'

use Analytics::Sinatra, <api_key>

before do
    content_type 'application/json'
end

get '/' do
    {message: 'Hello World!'}.to_json
end
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.

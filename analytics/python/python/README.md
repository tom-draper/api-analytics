# API Analytics

A free lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate an API key

Head to https://apianalytics.dev/generate to generate your unique API key with a single click. This key is used to monitor your API server and should be stored privately. It's also required in order to view your API analytics dashboard and data.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there is minimal impact on the performance of your API.

#### FastAPI

[![PyPi version](https://badgen.net/pypi/v/api-analytics)](https://pypi.com/project/api-analytics)

```bash
pip install fastapi-analytics
```

```py
import uvicorn
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, api_key=<API-KEY>)  # Add middleware

@app.get('/')
async def root():
    return {'message': 'Hello World!'}

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
```

##### Configuration

Custom mapping functions can be provided to override the default behaviour and tailor the retrival of information about each incoming request to your API environment and usage.

```py
from fastapi import FastAPI
from api_analytics.fastapi import Analytics, Config

config = Config()
config.get_ip_address = lambda request: request.headers['X-Forwarded-For']
config.get_user_agent = lambda request: request.headers['User-Agent']

app = FastAPI()
app.add_middleware(Analytics, api_key=<API-KEY>, config=config)  # Add middleware
```

#### Flask

[![PyPi version](https://badgen.net/pypi/v/api-analytics)](https://pypi.com/project/api-analytics)

```bash
pip install api-analytics
```

```py
from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <API-KEY>)  # Add middleware

@app.get('/')
def root():
    return {'message': 'Hello World!'}

if __name__ == "__main__":
    app.run()
```

##### Configuration

```py
from flask import Flask
from api_analytics.flask import add_middleware, Config

app = Flask(__name__)
config = Config()
config.get_ip_address = lambda request: request.headers['X-Forwarded-For']
config.get_user_agent = lambda request: request.headers['User-Agent']
add_middleware(app, <API-KEY>, config)  # Add middleware
```

#### Django

[![PyPi version](https://badgen.net/pypi/v/api-analytics)](https://pypi.com/project/api-analytics)

```bash
pip install api-analytics
```

Assign your API key to `ANALYTICS_API_KEY` in `settings.py` and add the Analytics middleware to the top of your middleware stack.

```py
ANALYTICS_API_KEY = <API-KEY>

MIDDLEWARE = [
    'api_analytics.django.Analytics',  # Add middleware
    ...
]
```

##### Configuration

Assign your config to `ANALYTICS_CONFIG` in `settings.py`.

```py
from api_analytics.django import Config

config = Config()
config.get_ip_address = lambda request: request.headers['X-Forwarded-For']
config.get_user_agent = lambda request: request.headers['User-Agent']
ANALYTICS_CONFIG = config
```

#### Tornado

[![PyPi version](https://badgen.net/pypi/v/api-analytics)](https://pypi.com/project/api-analytics)

```bash
pip install tornado-analytics
```

Modify your handler to inherit from `Analytics`. Create a `__init__()` method, passing along the application and response along with your unique API key.

```py
import asyncio
from tornado.web import Application
from api_analytics.tornado import Analytics

# Inherit from the Analytics middleware class
class MainHandler(Analytics):
    def __init__(self, app, res):
        super().__init__(app, res, <API-KEY>)  # Provide api key
    
    def get(self):
        self.write({'message': 'Hello World!'})

def make_app():
    return Application([
        (r"/", MainHandler),
    ])

if __name__ == "__main__":
    app = make_app()
    app.listen(8080)
    IOLoop.instance().start()
```

##### Configuration

```py
from api_analytics.tornado import Analytics, Config

class MainHandler(Analytics):
    def __init__(self, app, res):
        config = Config()
        config.get_ip_address = lambda request: request.headers['X-Forwarded-For']
        config.get_user_agent = lambda request: request.headers['User-Agent']
        super().__init__(app, res, <API-KEY>, config)  # Provide api key
```

### 3. View your analytics

Your API will now log and store incoming request data on all valid routes. Your logged data can be viewed using two methods:

1. Through visualizations and statistics on our dashboard
2. Accessed directly via our data API

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard. We recommend generating a new API key for each additional API server you want analytics for.

#### Dashboard

Head to https://apianalytics.dev/dashboard and paste in your API key to access your dashboard.

Demo: https://apianalytics.dev/dashboard/demo

![dashboard](https://user-images.githubusercontent.com/41476809/272061832-74ba4146-f4b3-4c05-b759-3946f4deb9de.png)

#### Data API

Logged data for all requests can be accessed via our REST API. Simply send a GET request to `https://apianalytics-server.com/api/data` with your API key set as `X-AUTH-TOKEN` in headers.

##### Python

```py
import requests

headers = {
 "X-AUTH-TOKEN": <API-KEY>
}

response = requests.get("https://apianalytics-server.com/api/data", headers=headers)
print(response.json())
```

##### Node.js

```js
fetch("https://apianalytics-server.com/api/data", {
  headers: { "X-AUTH-TOKEN": <API-KEY> },
})
  .then((response) => {
    return response.json();
  })
  .then((data) => {
    console.log(data);
  });
```

##### cURL

```bash
curl --header "X-AUTH-TOKEN: <API-KEY>" https://apianalytics-server.com/api/data
```

##### Parameters

You can filter your data by providing URL parameters in your request.

- `date` - specifies a particular day the requests occurred on (`YYYY-MM-DD`)
- `dateFrom` - specifies the lower bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `dateTo` - specifies the upper bound of a date range the requests occurred in (`YYYY-MM-DD`)
- `ipAddress` - an IP address string of the client
- `status` - an integer status code of the response
- `location` - a two-character location code of the client

Example:

```bash
curl --header "X-AUTH-TOKEN: <API-KEY>" https://apianalytics-server.com/api/data?dateFrom=2022-01-01&dateTo=2022-06-01&status=200
```

## Monitoring

Active API monitoring can be set up by heading to https://apianalytics.dev/monitoring to enter you API key. Our servers will regularly ping chosen API endpoints to monitor uptime and response time. 
<!-- Optional email alerts when your endpoints are down can be subscribed to. -->

![Monitoring](https://user-images.githubusercontent.com/41476809/208298759-f937b668-2d86-43a2-b615-6b7f0b2bc20c.png)

## Data and Security

All data is stored securely in compliance with The EU General Data Protection Regulation (GDPR).

For any given request to your API, data recorded is limited to:
 - Path requested by client
 - Client IP address
 - Client operating system
 - Client browser
 - Request method (GET, POST, PUT, etc.)
 - Time of request
 - Status code
 - Response time
 - API hostname
 - API framework (FastAPI, Flask, Express etc.)

Data collected is only ever used to populate your analytics dashboard. All stored data is anonymous, with the API key the only link between you and your logged request data. Should you lose your API key, you will have no method to access your API analytics.

### Data Deletion

At any time you can delete all stored data associated with your API key by going to https://apianalytics.dev/delete and entering your API key.

API keys and their associated logged request data are scheduled to be deleted after 6 months of inactivity.

## Contributions

Contributions, issues and feature requests are welcome.

- Fork it (https://github.com/tom-draper/api-analytics)
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

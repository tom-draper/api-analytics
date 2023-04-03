# Python API Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
pip install api-analytics
```

#### Django

Assign your API key to `ANALYTICS_API_KEY` in `settings.py` and add the Analytics middleware to the top of your middleware stack.

```py
ANALYTICS_API_KEY = <API-KEY>

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]
```

#### FastAPI

```py
import uvicorn
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, <API-KEY>)  # Add middleware

@app.get("/")
async def root():
    return {"message": "Hello World"}

if __name__ == "__main__":
    uvicorn.run("app:app", reload=True)
```

#### Flask

```py
from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <API-KEY>)  # Add middleware

@app.get("/")
def root():
    return {"message": "Hello World"}

if __name__ == "__main__":
    app.run()
```

#### Tornado

Modify your handler to inherit from `Analytics`. Create a `__init__()` method on your handler, passing along the application and response along with your unique API key.

```py
import asyncio
from tornado.web import Application

from api_analytics.tornado import Analytics

# Inherit from the Analytics middleware class
class MainHandler(Analytics):
    def __init__(self, app, res):
        api_key = os.environ.get("API_KEY")
        super().__init__(app, res, api_key)

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

### 3. View your analytics

Your API will now log and store incoming request data on all valid routes. Your logged data can be viewed using two methods:

1. Through visualizations and statistics on our dashboard
2. Accessed directly via our data API

You can use the same API key across multiple APIs, but all your data will appear in the same dashboard. We recommend generating a new API key for each additional API server you want analytics for.

#### Dashboard

Head to https://my-api-analytics.vercel.app/dashboard and paste in your API key to access your dashboard.

Demo: https://my-api-analytics.vercel.app/dashboard/demo

![Dashboard](https://user-images.githubusercontent.com/41476809/211800529-a84a0aa3-70c9-47d4-aa0d-7f9bbd3bc9b5.png)

#### Data API

Logged data for all requests can be accessed via our REST API. Simply send a GET request to `https://api-analytics-server.vercel.app/api/data` with your API key set as `X-AUTH-TOKEN` in headers.

##### Python

```py
import requests

headers = {
 "X-AUTH-TOKEN": <API-KEY>
}

response = requests.get("https://api-analytics-server.vercel.app/api/data", headers=headers)
print(response.json())
```
##### Node.js

```js
fetch("https://api-analytics-server.vercel.app/api/data", {
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
curl --header "X-AUTH-TOKEN: <API-KEY>" https://api-analytics-server.vercel.app/api/data
```

## Monitoring (coming soon)

Opt-in active API monitoring is coming soon. Our servers will regularly ping your API endpoints to monitor uptime and response time. Optional email alerts to notify you when your endpoints are down can be subscribed to.

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

Data collected is only ever used to populate your analytics dashboard. All data stored is anonymous, with the API key the only link between you and your logged request data. Should you lose your API key, you will have no method to access your API analytics.

### Delete Data

At any time, you can delete all stored data associated with your API key by going to https://my-api-analytics.vercel.app/delete and entering your API key.

API keys and their associated API request data are scheduled be deleted after 1 year of inactivity.

## Development

This project is still in the early stages of development and bugs are to be expected.

## Contributions

Contributions, issues and feature requests are welcome.

- Fork it (https://github.com/tom-draper/api-analytics)
- Create your feature branch (`git checkout -b my-new-feature`)
- Commit your changes (`git commit -am 'Add some feature'`)
- Push to the branch (`git push origin my-new-feature`)
- Create a new Pull Request

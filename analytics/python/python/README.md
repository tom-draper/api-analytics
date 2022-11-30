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
ANALYTICS_API_KEY = <api_key>

MIDDLEWARE = [
    'api_analytics.django.Analytics',  # Add middleware
    ...
]
```

#### FastAPI

```py
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, <api_key>)  # Add middleware

@app.get("/")
async def root():
    return {"message": "Hello World"}
```

#### Flask

```py
from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)  # Add middleware

@app.get("/")
def root():
    return {"message": "Hello World"}
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
        super().__init__(app, res, <api_key>)  # Pass api key

    def get(self):
        self.write({'message': 'Hello World!'})

def make_app():
    return Application([
        (r"/", MainHandler),
    ])

async def main():
    app = make_app()
    app.listen(8080)
    await asyncio.Event().wait()

if __name__ == "__main__":
    asyncio.run(main())
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.

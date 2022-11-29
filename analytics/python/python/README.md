# Python API Analytics

A lightweight API analytics solution, complete with a dashboard.

## Getting Started

### 1. Generate a new API key

Head to https://my-api-analytics.vercel.app/generate to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics dashboard.

### 2. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance.

```bash
python -m pip install api-analytics
```

#### Django

Set you API key as an environment variable.

In `settings.py`:

```py
from os import getenv

ANALYTICS_API_KEY = getenv("API_KEY")

MIDDLEWARE = [
    'api_analytics.django.Analytics',
    ...
]
```

#### FastAPI

```py
from fastapi import FastAPI
from api_analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, <api_key>)

@app.get("/")
async def root():
    return {"message": "Hello World"}
```

#### Flask

```py
from flask import Flask
from api_analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)

@app.get("/")
def root():
    return {"message": "Hello World"}
```

### 3. View your analytics

Your API will log requests on all valid routes. Head over to https://my-api-analytics.vercel.app/dashboard and paste in your API key to view your dashboard.

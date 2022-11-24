# FastAPI Analytics

A lightweight API analytics solution, complete with a statistics dashboard.

Currently available for FastAPI and Flask.

## Getting Started

### 1. Install pip package

```
python -m pip install api-analytics
```

### 2. Generate a new API key

Head to <URL> to generate your unique API key with a single click. This key is used to monitor your specific API, so keep it secret! It's also required in order to view your APIs analytics.

### 3. Add middleware to your API

Add our lightweight middleware to your API. Almost all processing is handled by our servers so there should be virtually no impact on your APIs performance. 

#### FastAPI

```py
from fastapi import FastAPI
from analytics.fastapi import Analytics

app = FastAPI()
app.add_middleware(Analytics, <api_key>)

@app.get("/")
async def root():
    return {"message": "Hello World"}
```

#### Flask

```py
from flask import Flask
from analytics.flask import add_middleware

app = Flask(__name__)
add_middleware(app, <api_key>)


@app.get("/")
def root():
    return {"message": "Hello World"}
```

### 4. View your analytics

Your API will log . Head over to <URL> and paste in your API key to view your dashboard.
